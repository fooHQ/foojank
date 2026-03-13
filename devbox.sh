#!/usr/bin/env bash

set -euo pipefail

build_foojank_dev() {
    OUTPUT="${OUTPUT:-build/foojank}"
    export CGO_ENABLED=1
    go build -race -tags dev -o "$OUTPUT" ./cmd/foojank
}

build_foojank_prod() {
    OUTPUT="${OUTPUT:-build/foojank}"
    go build -tags prod -o "$OUTPUT" ./cmd/foojank
}

test() {
    CGO_ENABLED=1 go test -race -timeout 30s -tags dev ./...
}

lint() {
    golangci-lint run --timeout 10m
}

eval $@
