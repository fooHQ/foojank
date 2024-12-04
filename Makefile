.PHONY: build/vessel/prod
build/vessel/prod:
	devbox run -e "CGO_ENABLED=0" -- go build -ldflags "-w -s" -o build/vessel ./cmd/vessel

.PHONY: build/vessel/small
build/vessel/small: build/vessel/prod
	devbox run -- upx --lzma build/vessel

.PHONY: build/vessel/dev
build/vessel/dev:
	devbox run -e "CGO_ENABLED=0" -- go build -tags debug -o build/vessel ./cmd/vessel

.PHONY: build/client/prod
build/client/prod:
	devbox run -e "CGO_ENABLED=0" -- go build -ldflags "-w -s" -o build/client ./cmd/client

.PHONY: build/client/dev
build/client/dev:
	devbox run -e "CGO_ENABLED=0" -- go build -tags debug -o build/client ./cmd/client

.PHONY: build/server/prod
build/server/prod:
	devbox run -e "CGO_ENABLED=0" -- go build -ldflags "-w -s" -o build/server ./cmd/server

.PHONY: build/server/dev
build/server/dev:
	devbox run -e "CGO_ENABLED=0" -- go build -tags debug -o build/server ./cmd/server

.PHONY: generate
generate:
	devbox run -- go generate ./...

.PHONY: run/vessel
run/vessel:
	devbox run -e "CGO_ENABLED=1" -- go run -race -tags debug ./cmd/vessel

.PHONY: run/server
run/server:
	devbox run -e "CGO_ENABLED=1" -- go run -race ./cmd/server start

.PHONY: test
test:
	devbox run -e "CGO_ENABLED=1" -- go test -race ./...

.PHONY: docker/dev
docker/dev:
	docker-compose -f docker/docker-compose.yaml up --force-recreate

.PHONY: docs/dev
docs/dev:
	docsify serve docs/

.PHONY: lint
lint:
	devbox run -- golangci-lint run
