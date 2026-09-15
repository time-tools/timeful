# Module-cache-free base. CI test images (compose.test.yaml) stop here: the
# test containers mount the source and a cached GOMODCACHE volume at
# /go/pkg/mod and compile live, so baking the module download cache into the
# image would only bloat every cached layer for discarded work.
FROM golang:1.26.4-alpine3.24 AS testbase

WORKDIR /app

# Backup/restore rehearsal runner. Only server-route-test uses this target; it
# adds the PostgreSQL client tools so the rehearsal can exercise a real
# pg_dump/pg_restore cycle against the isolated test database.
FROM testbase AS testbase-pg

RUN apk add --no-cache postgresql18-client

# Module-download stage for full image builds (builder chain). CI test images
# stop at testbase; the full server binary build below would be discarded
# work for them.
FROM testbase AS testdeps

COPY go.mod go.sum ./
RUN go mod download

# Build stage
FROM testdeps AS builder

WORKDIR /app

# Copy the remaining build inputs and compile the server binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /out/server .

# Goose runs schema changes in its own one-shot Compose service. Clean caches
# in the same RUN so the layer keeps only the goose binary; the goose module
# dependency cache (~hundreds of MB) must not land in the cached image layer.
FROM golang:1.26.4-alpine3.24 AS goose-builder
RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.26.0 \
	&& go clean -cache -modcache

FROM alpine:3.24 AS migrator
RUN apk add --no-cache postgresql18-client
COPY --from=goose-builder /go/bin/goose /usr/local/bin/goose
COPY migrations /migrations
CMD ["goose"]

# Runtime stage
FROM alpine:3.24

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user and writable runtime directories
RUN adduser -D -g '' appuser \
	&& mkdir -p /app/logs \
	&& chown -R appuser:appuser /app

# Copy binary from builder
COPY --from=builder /out/server ./server

USER appuser

EXPOSE 3002

CMD ["./server"]
