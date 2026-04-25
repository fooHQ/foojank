#!/usr/bin/env bash

set -euo pipefail

build() {
    OUTPUT="${OUTPUT:-build/foojank}"
    go build -o "$OUTPUT" ./cmd/foojank
}

test() {
    CGO_ENABLED=1 go test -race -timeout 30s -tags dev ./...
}

lint() {
    golangci-lint run --timeout 10m
}

eval $@
