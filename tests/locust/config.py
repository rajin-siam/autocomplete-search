# Locust Load Testing Configuration
# This file contains configurations for different load testing scenarios

# Test scenarios configuration
TEST_SCENARIOS = {
    "smoke_test": {
        "users": 10,
        "spawn_rate": 2,
        "duration": "30s",
        "description": "Basic smoke test with low load"
    },
    "normal_load": {
        "users": 50,
        "spawn_rate": 5,
        "duration": "2m",
        "description": "Simulate normal user traffic"
    },
    "high_load": {
        "users": 200,
        "spawn_rate": 20,
        "duration": "5m",
        "description": "Simulate peak traffic conditions"
    },
    "stress_test": {
        "users": 500,
        "spawn_rate": 50,
        "duration": "10m",
        "description": "Stress test to find breaking points"
    },
    "spike_test": {
        "users": 1000,
        "spawn_rate": 200,
        "duration": "2m",
        "description": "Spike test to handle sudden traffic bursts"
    }
}

# Target configuration
TARGET_HOST = "http://localhost:2322"  # Your place-search API

# Test data
SEARCH_QUERIES = [
    # Exact matches
    "Dhaka", "Dinajpur", "London", "Paris", "Tokyo",
    
    # Fuzzy matches (with typos)
    "dhka", "lndon", "pari", "toky", "lndon",
    
    # Prefix matches
    "dh", "lo", "pa", "to", "ne", "ca", "ny", "sy",
    
    # No results
    "nonexistent", "xyzabc", "nowhereland",
    
    # Special characters
    "Cox's Bazar", "O'Connell Street", "St. Louis",
    
    # Unicode/Bengali
    "ঢাকা", "চট্টগ্রাম", "খুলনা", "বরিশাল"
]

# Performance thresholds (for monitoring)
THRESHOLDS = {
    "response_time_95th_percentile": 500,  # ms
    "response_time_avg": 200,  # ms
    "error_rate": 0.01,  # 1%
    "requests_per_second": 1000
}