#!/bin/bash

set -e

echo "🔧 Preparing test environment..."

mkdir -p internal/static

# static.go dosyasını üret
cat > internal/static/static.go <<EOF
package static

import "embed"

//go:embed *
var Static embed.FS
EOF

echo "Running unit tests (go test)..."
go test -tags=test ./internal/...

echo "Running Ginkgo tests..."
ginkgo -tags=test -r ./internal/...