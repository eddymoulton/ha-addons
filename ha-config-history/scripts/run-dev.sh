#!/usr/bin/env bash

mkdir -p tmp/data

./scripts/build-frontend.sh && \
CONFIG_PATH="tmp/data/config.json" go run ./main.go