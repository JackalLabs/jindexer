# syntax=docker/dockerfile:1.6

FROM golang:1.25.3-bookworm AS builder

WORKDIR /app

# Install build dependencies including CGO requirements for SQLite
# Note: Using Debian-based image for GLIBC compatibility with wasmvm
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies and copy wasmvm libraries in one step
# We need to access the module cache to copy the wasmvm .so files
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /wasmvm-libs && \
    go mod download && \
    (find /go/pkg/mod -path "*wasmvm*" -name "*.so" -type f -exec cp {} /wasmvm-libs/ \; 2>/dev/null || true) && \
    (find /go/pkg/mod -path "*wasmvm*" -name "*.so.*" -type f -exec cp {} /wasmvm-libs/ \; 2>/dev/null || true)

COPY . .

# Build the binary with CGO enabled (required for go-sqlite3 and wasmvm)
# Note: CGO_ENABLED=1 is the default, but explicitly set for clarity
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o jindexer .

# Final stage
FROM debian:bookworm-slim AS final

# Install runtime dependencies: libsqlite3 for CGO SQLite driver
# ca-certificates for HTTPS connections
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libsqlite3-0 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

# Copy the binary from builder stage to /usr/local/bin so it's in PATH
# and won't be shadowed by volume mounts
COPY --from=builder /app/jindexer /usr/local/bin/jindexer

# Copy wasmvm shared libraries directory and install them
# The directory always exists (created in builder), but may be empty
COPY --from=builder /wasmvm-libs /tmp/wasmvm-libs
RUN if [ -d /tmp/wasmvm-libs ] && [ "$(ls -A /tmp/wasmvm-libs 2>/dev/null)" ]; then \
        cp /tmp/wasmvm-libs/*.so* /usr/lib/ 2>/dev/null || true; \
    fi && \
    rm -rf /tmp/wasmvm-libs

# Set library path to ensure wasmvm libraries can be found
ENV LD_LIBRARY_PATH=/usr/lib:${LD_LIBRARY_PATH}

CMD ["jindexer"]
