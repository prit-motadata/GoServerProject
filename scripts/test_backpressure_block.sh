#!/bin/bash

# Test Backpressure: BLOCK
# Sends 15 concurrent requests to /logs
# Expected: All 202 Accepted, but requests will wait for channel space.
# The server workers take 2 seconds per log, and queue size is 5.
# With 3 workers, it should process 3 at a time.
# Total capacity before blocking: 3 (workers) + 5 (queue) = 8.
# Requests 9-15 should block.

# IMPORTANT: Run server with high burst to avoid rate limiting:
# export RATE_LIMIT=100 && export RATE_BURST=100 && go run cmd/main.go

URL="http://localhost:8080/logs"
CONCURRENT_REQUESTS=15

echo "Sending $CONCURRENT_REQUESTS concurrent requests to $URL (Strategy: BLOCK)..."
echo "Note: Ensure server is running with high RATE_BURST and Strategy=BLOCK"

for i in $(seq 1 $CONCURRENT_REQUESTS); do
    payload="{\"level\":\"INFO\",\"service\":\"test-service\",\"message\":\"log-$i\",\"timestamp\":\"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\"}"
    curl -s -o /dev/null -w "Request $i: %{http_code} (time: %{time_total}s)\n" -X POST -d "$payload" "$URL" &
done

wait
echo "Done."
