#!/usr/bin/env bash

build_agent_dev() {
    if [ -z "$OUTPUT" ]; then
        echo "OUTPUT variable not defined"
        return 1
    fi

    WITH_RACE=""
    if [ "$GOOS" != "windows" ]; then
        WITH_RACE="-race"
        export CGO_ENABLED=1
    fi

    TAGS="dev $TAGS"
    go build -tags "$TAGS" -o "$OUTPUT" $WITH_RACE ./cmd/vessel
}

build_agent_prod() {
    if [ -z "$OUTPUT" ]; then
        echo "OUTPUT variable not defined"
        return 1
    fi

    WITH_LDFLAGS="-w -s"
    if [ "$GOOS" = "windows" ]; then
        WITH_LDFLAGS="$WITH_LDFLAGS -H windowsgui"
    fi

    # TODO: enable garble
    #garble -tiny build -tags prod -o "$OUTPUT" -ldflags="$WITH_LDFLAGS" ./cmd/vessel
    go build -tags "$TAGS" -o "$OUTPUT" -ldflags="$WITH_LDFLAGS" ./cmd/vessel
}

build_foojank_dev() {
    if [ -z "$OUTPUT" ]; then
        OUTPUT="build/foojank"
    fi

    export CGO_ENABLED=1
    go build -race -tags dev -o "$OUTPUT" ./cmd/foojank
}

build_foojank_prod() {
    if [ -z "$OUTPUT" ]; then
        OUTPUT="build/foojank"
    fi

    go build -tags prod -o "$OUTPUT" ./cmd/foojank
}

build_runscript() {
    if [ -z "$OUTPUT" ]; then
        echo "OUTPUT variable not defined"
        return 1
    fi

    TAGS="dev $TAGS"
    go build -tags "$TAGS" -o "$OUTPUT" ./cmd/runscript
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
