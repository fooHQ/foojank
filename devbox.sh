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

generate_proto() {
    if [ ! -d "./build/go-capnp" ]; then
        git clone -b v3.0.1-alpha.2 --depth 1 https://github.com/capnproto/go-capnp ./build/go-capnp
    fi

    cd ./build/go-capnp || exit 1
    go build -modfile go.mod -o ../capnpc-go ./capnpc-go
    cd - || exit 1
    go generate ./proto/capnp
}

eval $@
