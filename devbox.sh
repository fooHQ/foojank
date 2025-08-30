#!/usr/bin/env bash

set -euo pipefail

build_agent_dev() {
    WITH_LDFLAGS="$(agent_config)"
    WITH_RACE=""
    if [ "$GOOS" != "windows" ]; then
        WITH_RACE="-race"
        export CGO_ENABLED=1
    fi

    TAGS="dev $TAGS"
    go build -tags "$TAGS" -o "$OUTPUT" $WITH_RACE -ldflags "$WITH_LDFLAGS" ./cmd/vessel
}

build_agent_prod() {
    WITH_LDFLAGS="-w -s $(agent_config)"
    if [ "$GOOS" = "windows" ]; then
        WITH_LDFLAGS="$WITH_LDFLAGS -H windowsgui"
    fi

    # TODO: enable garble
    #garble -tiny build -tags prod -o "$OUTPUT" -ldflags="$WITH_LDFLAGS" ./cmd/vessel
    go build -tags "$TAGS" -o "$OUTPUT" -ldflags "$WITH_LDFLAGS" ./cmd/vessel
}

build_foojank_dev() {
    OUTPUT="${OUTPUT:-build/foojank}"
    export CGO_ENABLED=1
    go build -race -tags dev -o "$OUTPUT" ./cmd/foojank
}

build_foojank_prod() {
    OUTPUT="${OUTPUT:-build/foojank}"
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

agent_config() {
   cat <<EOF | tr '\n' ' '
-X main.ID=$ID
-X main.Server=$SERVER
-X main.UserJWT=$USER_JWT
-X main.UserKey=$USER_KEY
-X main.CACertificate=$CA_CERTIFICATE
-X main.Stream=$STREAM
-X main.Consumer=$CONSUMER
-X main.InboxPrefix=$INBOX_PREFIX
-X main.ObjectStoreName=$OBJECT_STORE_NAME
-X main.SubjectApiWorkerStartT=$SUBJECT_API_WORKER_START_T
-X main.SubjectApiWorkerStopT=$SUBJECT_API_WORKER_STOP_T
-X main.SubjectApiWorkerWriteStdinT=$SUBJECT_API_WORKER_WRITE_STDIN_T
-X main.SubjectApiWorkerWriteStdoutT=$SUBJECT_API_WORKER_WRITE_STDOUT_T
-X main.SubjectApiWorkerStatusT=$SUBJECT_API_WORKER_STATUS_T
-X main.SubjectApiConnInfoT=$SUBJECT_API_CONN_INFO_T
-X main.SubjectApiReplyT=$SUBJECT_API_REPLY_T
EOF
}

eval $@
