# This work is licensed under Apache License, Version 2.0 or later.
# Please read and understand latest version of Licence.

# stage 0: Build Frontend (Node.js)
FROM docker.io/library/node:23-alpine3.21 AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install

COPY frontend/ ./
RUN npm run build && \
    find dist -type f ! -name '*.gz' -delete

# Stage 1: Build Backend (Go)
FROM docker.io/library/golang:1.23-alpine3.21 AS backend-builder

ARG REV
ARG BUILD_TIME

ENV PROJECT=github.com/kazimsarikaya/assesmentbarkinrl

WORKDIR /app

# Copy backend code
COPY . .

# Copy built frontend files from stage 0
COPY --from=frontend-builder /app/frontend/dist/ ./internal/static/

# Generate embed static.go
RUN cat > ./internal/static/static.go <<EOF
package static

import "embed"

//go:embed *
var Static embed.FS
EOF

# Tidy up and vendor deps
RUN go mod tidy && go mod vendor

# Build binary with version info
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w -X '$PROJECT/internal/config.version=$REV' -X '$PROJECT/internal/config.buildTime=$BUILD_TIME' -X '$PROJECT/internal/config.goVersion=$(go version)'" -o /go_react_mui ./cmd

# Stage 2: Final scratch container
FROM scratch AS final

env HOME = /config

# Copy binary
COPY --from=backend-builder /go_react_mui /go_react_mui

# Copy any required certs or config (optional, depending on your app)
COPY --from=backend-builder /etc/ssl/certs /etc/ssl/certs

ENTRYPOINT ["/go_react_mui"]

