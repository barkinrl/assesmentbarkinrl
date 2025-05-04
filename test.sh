#!/bin/bash

set -e

echo "Running unit tests (go test)..."
go test -tags=test ./internal/...

echo "Running Ginkgo tests..."
ginkgo -tags=test -r ./internal/...