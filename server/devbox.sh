#!/usr/bin/env bash

set -euo pipefail

STORE_DIR="./nsc"
AUTH_FILE="auth.conf"

setup() {
    nsc env --store "$STORE_DIR"
    nsc add operator --sys --name "foojank-dummy-operator"
    nsc edit operator --account-jwt-server-url "nats://127.0.0.1"
    nsc generate config --nats-resolver > "$AUTH_FILE"
}

start() {
    nats-server -c "server.conf"
}

import() {
    if [ $# -eq 0 ]; then
        echo "Usage: $0 import <jwt>"
        exit 1
    fi
    local tmp="$(mktemp).jwt"
    echo "$1" > "$tmp"
    nsc import account --file "$tmp" --overwrite
    nsc push -P -A
}

clean() {
    rm -rf "$STORE_DIR"
    rm -rf "./jwt"
}

eval $@
