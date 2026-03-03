#!/bin/bash

# Test Rate Limiting
# Sends 10 concurrent requests to /health
# Expected: Some 200 OK, some 429 Too Many Requests (if burst is exceeded)

URL="http://localhost:8080/health"
CONCURRENT_REQUESTS=10

echo "Sending $CONCURRENT_REQUESTS concurrent requests to $URL..."

for i in $(seq 1 $CONCURRENT_REQUESTS); do
    curl -s -o /dev/null -w "Request $i: %{http_code}\n" "$URL" &
done

wait
echo "Done."
