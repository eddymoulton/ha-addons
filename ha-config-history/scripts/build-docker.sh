#!/bin/bash
# Docker build script with version information

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')}"
GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo 'none')}"

echo "Building Docker image with:"
echo "  VERSION:    $VERSION"
echo "  GIT_COMMIT: $GIT_COMMIT"

docker build \
  --build-arg VERSION="$VERSION" \
  --build-arg GIT_COMMIT="$GIT_COMMIT" \
  -t ha-config-history:latest \
  -t ha-config-history:$VERSION \
  .
