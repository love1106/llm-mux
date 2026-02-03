#!/bin/bash
set -e

cd "$(dirname "$0")/.."

echo "Building llm-mux..."
/usr/local/go/bin/go build -o llm-mux ./cmd/server

echo "Build complete: ./llm-mux"
