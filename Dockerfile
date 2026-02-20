# Build stage: compile a static Go binary
FROM golang:1.26-bookworm AS builder

WORKDIR /build

# Cache dependency downloads in a separate layer
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code (only what's needed for compilation)
COPY cmd/ cmd/
COPY internal/ internal/

# Build metadata injected via build args (defaults for local builds)
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

# Build a fully static binary with debug info stripped
RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w \
      -X github.com/jarfernandez/check-image/internal/version.Version=v${VERSION} \
      -X github.com/jarfernandez/check-image/internal/version.Commit=${COMMIT} \
      -X github.com/jarfernandez/check-image/internal/version.BuildDate=${BUILD_DATE}" \
    -o /check-image \
    ./cmd/check-image

# Final stage: minimal distroless image with CA certificates and non-root user
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /check-image /check-image

USER nonroot:nonroot

ENTRYPOINT ["/check-image"]
