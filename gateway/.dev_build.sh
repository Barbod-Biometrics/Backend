#!/bin/sh
set -e

# Generate Swagger docs if swag is available; do not fail the build if swag is missing.
if command -v swag >/dev/null 2>&1; then
  echo "swag found: generating swagger docs..."
  # point the generator at the main package and write into ./docs (existing folder)
  swag init -g ./cmd/app/main.go -o ./docs || true
else
  echo "swag not found: skipping swagger generation"
fi

if command -v wire >/dev/null 2>&1; then
  echo "wire found: generating dependency injection code..."
  # Generate dependency injection code
  wire ./wire
else
  echo "wire not found: skipping dependency injection code generation"
fi

echo "building binary..."
go build -ldflags='-w -s' -o ./tmp/main ./cmd/app

echo "build complete: ./tmp/main"
