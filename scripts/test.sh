#!/bin/bash

# Run tests with coverage
go test ./... -coverprofile=coverage.out

# Display coverage report
go tool cover -func=coverage.out

# Cleanup if requested
if [ "$1" == "--clean" ]; then
    rm coverage.out
fi
