#!/bin/bash

# Test Backpressure: DROP
# Sends 20 concurrent requests to /logs
# Expected: All 202 Accepted, but many will be dropped silently by the server.

URL="http://localhost:8080/logs"
CONCURRENT_REQUESTS=20

echo "Sending $CONCURRENT_REQUESTS concurrent requests to $URL (Strategy: DROP)..."

for i in $(seq 1 $CONCURRENT_REQUESTS); do
    payload="{\"level\":\"INFO\",\"service\":\"test-service\",\"message\":\"log-$i\",\"timestamp\":\"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\"}"
    curl -s -o /dev/null -w "Request $i: %{http_code}\n" -X POST -d "$payload" "$URL" &
done

wait
echo "Done. Check server logs to see dropped entries."
