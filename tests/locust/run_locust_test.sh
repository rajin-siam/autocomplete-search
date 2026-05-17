#!/bin/bash

# Load test runner script for Place Search API
# Usage: ./run_locust_test.sh [scenario]

# Set default scenario if not provided
SCENARIO=${1:-normal_load}

# Target host
TARGET_HOST="http://localhost:2322"

# Define scenarios
case $SCENARIO in
    smoke_test)
        USER_COUNT=10
        SPAWN_RATE=2
        DURATION="30s"
        DESCRIPTION="Basic smoke test with low load"
        ;;
    normal_load)
        USER_COUNT=50
        SPAWN_RATE=5
        DURATION="2m"
        DESCRIPTION="Simulate normal user traffic"
        ;;
    high_load)
        USER_COUNT=200
        SPAWN_RATE=20
        DURATION="5m"
        DESCRIPTION="Simulate peak traffic conditions"
        ;;
    stress_test)
        USER_COUNT=500
        SPAWN_RATE=50
        DURATION="10m"
        DESCRIPTION="Stress test to find breaking points"
        ;;
    spike_test)
        USER_COUNT=1000
        SPAWN_RATE=200
        DURATION="2m"
        DESCRIPTION="Spike test to handle sudden traffic bursts"
        ;;
    *)
        echo "Error: Scenario '$SCENARIO' not found"
        echo "Available scenarios:"
        echo "  smoke_test: Basic smoke test with low load"
        echo "  normal_load: Simulate normal user traffic"
        echo "  high_load: Simulate peak traffic conditions"
        echo "  stress_test: Stress test to find breaking points"
        echo "  spike_test: Spike test to handle sudden traffic bursts"
        exit 1
        ;;
esac

echo "=== Starting Locust Load Test ==="
echo "Scenario: $SCENARIO"
echo "Description: $DESCRIPTION"
echo "Users: $USER_COUNT"
echo "Spawn Rate: $SPAWN_RATE users/second"
echo "Duration: $DURATION"
echo "Target: $TARGET_HOST"
echo ""

# Check if locust is installed
if ! command -v locust &> /dev/null; then
    echo "Error: Locust is not installed"
    echo "Install it with: pip install locust"
    exit 1
fi

# Check if server is running
if ! curl -s "$TARGET_HOST/api?q=test" > /dev/null 2>&1; then
    echo "Warning: Target server at $TARGET_HOST may not be responding"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Run the test
echo "Starting load test..."
echo "Open http://localhost:8089 to monitor the test (if running with web UI)"
echo ""

cd "$(dirname "$0")"
locust -f locustfile.py \
  --host="$TARGET_HOST" \
  --users="$USER_COUNT" \
  --spawn-rate="$SPAWN_RATE" \
  --run-time="$DURATION" \
  --html=report_"$SCENARIO".html \
  --csv=stats_"$SCENARIO" \
  --headless

echo ""
echo "=== Test Complete ==="
echo "Report generated: report_$SCENARIO.html"
echo "Stats saved to: stats_$SCENARIO.csv"