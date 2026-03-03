#!/bin/bash

# Test Backpressure: REJECT
# Sends 15 concurrent requests to /logs
# Expected: Some 202 Accepted, some 429 Too Many Requests when queue is full.

URL="http://localhost:8080/logs"
CONCURRENT_REQUESTS=15

echo "Sending $CONCURRENT_REQUESTS concurrent requests to $URL (Strategy: REJECT)..."

for i in $(seq 1 $CONCURRENT_REQUESTS); do
    payload="{\"level\":\"INFO\",\"service\":\"test-service\",\"message\":\"log-$i\",\"timestamp\":\"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\"}"
    curl -s -o /dev/null -w "Request $i: %{http_code}\n" -X POST -d "$payload" "$URL" &
done

wait
echo "Done."
