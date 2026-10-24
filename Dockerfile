# Build stage
FROM golang:1.27-alpine AS builder

ARG VERSION=v0.0.0
ARG COMMIT_HASH=0000000
ARG BUILD_TIMESTAMP=1970-01-01T00:00:00
ARG PROJECT_URL=https://github.com/babs/runstats

WORKDIR /build

COPY go.mod ./
COPY main.go ./

RUN BUILDER="$(go version)" && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X 'main.Version=${VERSION}' \
    -X 'main.CommitHash=${COMMIT_HASH}' \
    -X 'main.BuildTimestamp=${BUILD_TIMESTAMP}' \
    -X 'main.Builder=${BUILDER}' \
    -X 'main.ProjectURL=${PROJECT_URL}'" \
    -o runstats .

# Runtime stage
FROM alpine:latest

ARG VERSION=v0.0.0
ARG COMMIT_HASH=0000000
ARG BUILD_TIMESTAMP=1970-01-01T00:00:00
ARG PROJECT_URL=https://github.com/babs/runstats

LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${COMMIT_HASH}"
LABEL org.opencontainers.image.created="${BUILD_TIMESTAMP}"
LABEL org.opencontainers.image.url="${PROJECT_URL}"
LABEL org.opencontainers.image.source="${PROJECT_URL}"

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/runstats /app/runstats

ENTRYPOINT ["/app/runstats"]
