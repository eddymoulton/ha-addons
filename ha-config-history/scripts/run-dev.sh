#!/usr/bin/env bash

mkdir -p tmp/data

./scripts/build-frontend.sh && \
APPSETTINGS_PATH="tmp/data/appsettings.json" go run ./main.go