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

```
├── cmd/
│   ├── place-search/      # Application entry point
│   ├── root.go            # CLI root command (cobra)
│   ├── serve.go           # HTTP server command
│   └── index.go           # Data indexing command
├── config/                # Configuration loading (viper)
├── internal/
│   ├── domain/            # Business entities
│   ├── application/       # Use cases and services
│   ├── infrastructure/    # Elasticsearch client and repository
│   └── interfaces/        # HTTP handlers and GeoJSON formatter
├── config.yaml            # Default configuration
└── docker-compose.yml
```

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL with Nominatim data (for indexing)

## Quick Start

1. **Clone and configure**
   ```bash
   cp config.yaml.example config.yaml  # edit with your settings
   ```

2. **Start Elasticsearch**
   ```bash
   docker-compose up -d
   ```

3. **Index data** (requires a running Nominatim PostgreSQL database)
   ```bash
   go run cmd/place-search/main.go index
   ```

4. **Start the server**
   ```bash
   go run cmd/place-search/main.go serve
   ```

## Configuration

All settings live in `config.yaml`. Sensitive values can be overridden via environment variables.

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
  password: ""   # override with DATABASE_PASSWORD env var
```

### Environment variable overrides

| Variable            | Config key               |
|---------------------|--------------------------|
| `ES_HOST`           | `elasticsearch.host`     |
| `ES_INDEX`          | `elasticsearch.index`    |
| `SERVER_PORT`       | `server.port`            |
| `SEARCH_MIN_CHARS`  | `search.min_chars`       |
| `SEARCH_FUZZINESS`  | `search.fuzziness`       |
| `SEARCH_MAX_RESULTS`| `search.max_results`     |
| `DATABASE_PASSWORD` | `database.password`      |

## API

**Endpoint:** `GET /search?q={query}`

```bash
curl "http://localhost:2322/search?q=dhka"
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
# run tests
go test ./...

# build binary
go build -o place-search cmd/place-search/main.go
```

## Data Indexing

The `index` command connects to a Nominatim PostgreSQL database, extracts place data, creates the Elasticsearch index with a `search_as_you_type` mapping on the `name` field, and bulk-indexes all documents.

Configure the database connection in `config.yaml` (or via `DATABASE_PASSWORD` env var) before running.

## License

MIT
