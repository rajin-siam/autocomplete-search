# Place Search

A high-performance geocoding search service built with Go and Elasticsearch, providing fast place name lookups with fuzzy matching and GeoJSON output.

## Features

- **Search-as-you-type** using Elasticsearch's `search_as_you_type` field type
- **Typo-tolerant fuzzy matching** — handles misspellings like `dhka` → `Dhaka`
- **GeoJSON output** for easy integration with mapping libraries
- **Configurable search parameters** (minimum characters, fuzziness, result limits)
- **Clean architecture** with domain-driven design
- **CLI interface** with `serve` and `index` subcommands
- **Docker support** for easy deployment

## Architecture

This project follows **Clean Architecture** with **SOLID principles** and **Dependency Inversion**.

```
├── cmd/
│   ├── root.go            # CLI root command (cobra)
│   ├── serve.go           # HTTP server command
│   └── index.go           # Data indexing command (refactored)
├── config/                # Configuration loading (viper + godotenv)
├── internal/
│   ├── domain/            # Business entities (Place, IndexDocument)
│   ├── application/       # Use cases, services, and interfaces
│   │   ├── indexing_service.go      # Orchestrates indexing workflow
│   │   ├── search_service.go        # Search business logic
│   │   ├── place_extractor.go       # Interface for data extraction
│   │   ├── index_manager.go         # Interfaces for index management
│   │   └── document_indexer.go      # Interface for bulk indexing
│   ├── infrastructure/    # Concrete implementations
│   │   ├── elasticsearch/ # ES client, index manager, bulk indexer
│   │   ├── nominatim/     # PostgreSQL data extraction, HStore parsing
│   │   └── hstore/        # PostgreSQL HStore parser
│   └── interfaces/        # HTTP handlers and GeoJSON formatter
├── config.yaml            # Default configuration
├── .env                   # Environment variables (auto-loaded)
└── docker-compose.yml     # Full stack deployment
```

### Key Design Patterns

- **Dependency Injection**: All dependencies injected via constructors
- **Interface Segregation**: Small, focused interfaces (PlaceExtractor, IndexManager, DocumentIndexer)
- **Repository Pattern**: Database access abstracted behind interfaces
- **Service Layer**: Business logic separated from infrastructure
- **Streaming**: Channel-based data flow for memory efficiency

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL with Nominatim data (for indexing)
- Bangladesh OSM data file: `bangladesh-latest.osm.pbf` (for full setup)

## Quick Start

### Option 1: Full Stack with Docker Compose (Recommended)

```bash
# 1. Download Bangladesh OSM data
wget https://download.geofabrik.de/asia/bangladesh-latest.osm.pbf

# 2. Setup environment variables
cp .env.example .env
nano .env  # Set NOMINATIM_PASSWORD=your_password

# 3. Start all services (Nominatim, Elasticsearch, Indexer, API)
docker-compose up -d

# 4. Wait for indexing to complete (30-60 mins for first run)
docker-compose logs -f indexer

# 5. Test the API
curl "http://localhost:2322/api?q=dhaka"
```

### Option 2: Local Development

```bash
# 1. Start dependencies
docker-compose up -d nominatim elasticsearch

# 2. Setup environment (password auto-loaded from .env)
cp .env.example .env
nano .env  # Set NOMINATIM_PASSWORD=your_password

# 3. Install Go dependencies
go mod download

# 4. Index data
go run main.go index

# 5. Start server
go run main.go serve

# 6. Test
curl "http://localhost:2322/api?q=dhaka"
```

### Option 3: Quick Test (Without Nominatim)

```bash
# 1. Start only Elasticsearch
docker-compose up -d elasticsearch

# 2. Manually add test data
curl -X PUT "localhost:9200/places" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "name": {"type": "search_as_you_type"},
      "coordinate": {"type": "geo_point", "index": false}
    }
  }
}'

curl -X POST "localhost:9200/places/_doc" -H 'Content-Type: application/json' -d'
{
  "name": "Dhaka",
  "coordinate": {"lat": 23.7104, "lon": 90.4074},
  "osm_id": 123456,
  "osm_type": "relation",
  "type": "city",
  "country": "Bangladesh",
  "countrycode": "BD"
}'

# 3. Start server
go run main.go serve

# 4. Test
curl "http://localhost:2322/api?q=dhaka"
```

## Configuration

Configuration is loaded from multiple sources (in order of priority):

1. **Environment variables** (highest priority)
2. **`.env` file** (auto-loaded via godotenv)
3. **`config.yaml`** (default values)

### config.yaml

```yaml
elasticsearch:
  host: http://localhost:9200
  index: places

server:
  port: 2322

search:
  min_chars: 2
  fuzziness: AUTO
  max_results: 15

database:
  host: localhost
  port: 5432
  name: nominatim
  user: nominatim
  password: ""   # Leave empty, use .env file
```

### .env file (Recommended for secrets)

```bash
# Create from example
cp .env.example .env

# Edit and add your password
NOMINATIM_PASSWORD=your_secure_password
```

**Note:** The `.env` file is automatically loaded by the application. No need to manually export variables!

### Environment variable overrides

| Variable            | Config key               | Example |
|---------------------|--------------------------|----------|
| `ES_HOST`           | `elasticsearch.host`     | `http://localhost:9200` |
| `ES_INDEX`          | `elasticsearch.index`    | `places` |
| `SERVER_PORT`       | `server.port`            | `2322` |
| `SEARCH_MIN_CHARS`  | `search.min_chars`       | `2` |
| `SEARCH_FUZZINESS`  | `search.fuzziness`       | `AUTO` |
| `SEARCH_MAX_RESULTS`| `search.max_results`     | `15` |
| `DATABASE_HOST`     | `database.host`          | `localhost` |
| `DATABASE_PORT`     | `database.port`          | `5432` |
| `DATABASE_NAME`     | `database.name`          | `nominatim` |
| `DATABASE_USER`     | `database.user`          | `nominatim` |
| `DATABASE_PASSWORD` | `database.password`      | `your_password` (use .env) |

## API

**Endpoint:** `GET /api?q={query}`

```bash
curl "http://localhost:2322/api?q=dhka"
```

**Response:**
```json
{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "geometry": {
        "type": "Point",
        "coordinates": [90.4074, 23.7104]
      },
      "properties": {
        "osm_id": 123456,
        "osm_type": "relation",
        "name": "Dhaka",
        "type": "city",
        "country": "Bangladesh",
        "countrycode": "BD"
      }
    }
  ]
}
```

**Error responses:**

| Status | Reason |
|--------|--------|
| `400`  | Missing `q` parameter or query shorter than `min_chars` |
| `500`  | Elasticsearch error |

## Search Behavior

Queries use a `bool/should` strategy combining two clauses:

- `multi_match bool_prefix` on `name` and its ngram subfields — handles autocomplete (e.g. `dha` → `Dhaka`)
- `match` with `fuzziness: AUTO` on `name` — handles typos (e.g. `dhka` → `Dhaka`)

Documents matching both clauses score higher, so exact/prefix matches naturally rank above fuzzy ones.

## Development

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build binary
go build -o place-search main.go

# Run with hot reload (install air first: go install github.com/cosmtrek/air@latest)
air

# Format code
go fmt ./...

# Lint code
golangci-lint run
```

## Monitoring & Debugging

```bash
# Check Elasticsearch health
curl http://localhost:9200/_cluster/health

# Check index stats
curl http://localhost:9200/places/_stats

# Count documents
curl http://localhost:9200/places/_count

# View sample documents
curl http://localhost:9200/places/_search?size=5

# Check Docker services
docker-compose ps

# View logs
docker-compose logs -f app
docker-compose logs -f elasticsearch
docker-compose logs -f nominatim
```

## Data Indexing

The `index` command follows a **clean architecture** approach:

### Indexing Workflow

1. **Index Creation** (`IndexManager`)
   - Deletes old index if exists
   - Creates new index with `search_as_you_type` mapping
   - Defines field types and indexing rules

2. **Data Extraction** (`PlaceExtractor`)
   - Queries Nominatim PostgreSQL database
   - Parses HStore format (multilingual names)
   - Extracts hierarchical address parts (state, city, district)
   - Transforms to domain model (`IndexDocument`)
   - Streams via channel (memory efficient)

3. **Bulk Indexing** (`DocumentIndexer`)
   - Batches documents (5MB chunks, 4 workers)
   - Sends to Elasticsearch in parallel
   - Tracks success/failure statistics
   - Handles individual document errors gracefully

### Running the Indexer

```bash
# Ensure .env file has DATABASE_PASSWORD
cat .env

# Run indexer
go run main.go index

# Expected output:
# Connected to PostgreSQL.
# Connected to Elasticsearch.
# Index created.
# Extracting and indexing...
# Indexed: 45230, Failed: 12
# Done.
```

### Indexing Architecture

The indexer uses **dependency injection** and **interfaces** for flexibility:

- `PlaceExtractor` interface → `PlaceExtractorImpl` (PostgreSQL)
- `IndexManager` interface → `IndexManagerImpl` (Elasticsearch)
- `DocumentIndexer` interface → `BulkDocumentIndexer` (Elasticsearch)
- `PlaceDatabase` interface → `PostgresPlaceDB` (PostgreSQL)

This design allows easy testing with mocks and swapping implementations (e.g., MySQL instead of PostgreSQL).

### HStore Parsing

Nominatim stores multilingual names in PostgreSQL HStore format:

```
"name"=>"Dhaka", "name:en"=>"Dhaka", "name:bn"=>"ঢাকা"
```

The indexer extracts names with priority: `name:en` → `name` → `name:bn` → any value.

### Configuration

Configure database connection in `.env` file:

```bash
DATABASE_PASSWORD=your_password
```

Or override in `config.yaml`:

```yaml
database:
  host: localhost
  port: 5432
  name: nominatim
  user: nominatim
  password: "your_password"  # Not recommended, use .env
```

## License

MIT
