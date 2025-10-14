#!/bin/bash
# Build script for runstats with version injection

set -e

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")}"
COMMIT_HASH="$(git rev-parse --short HEAD 2>/dev/null || echo "0000000")"
DIRTY=""
if ! git diff --quiet 2>/dev/null; then
    DIRTY="-dirty"
fi
BUILD_TIMESTAMP="${BUILD_TIMESTAMP:-$(date -u +"%Y-%m-%dT%H:%M:%S")}"
BUILDER="${BUILDER:-$(go version)}"
PROJECT_URL="${PROJECT_URL:-https://github.com/babs/runstats}"

echo "Building runstats ${VERSION}"
echo "  Commit: ${COMMIT_HASH}${DIRTY}"
echo "  Build time: ${BUILD_TIMESTAMP}"
echo "  Builder: ${BUILDER}"

LDFLAGS=(
    "-X 'main.Version=${VERSION}'"
    "-X 'main.CommitHash=${COMMIT_HASH}${DIRTY}'"
    "-X 'main.BuildTimestamp=${BUILD_TIMESTAMP}'"
    "-X 'main.Builder=${BUILDER}'"
    "-X 'main.ProjectURL=${PROJECT_URL}'"
)

go build -ldflags="${LDFLAGS[*]}" -o runstats .

echo "Build complete: ./runstats"
