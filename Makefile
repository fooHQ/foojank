export CGO_ENABLED=0

GO=devbox run -- go
UPX=devbox run -- upx
LINTER=devbox run -- golangci-lint

.PHONY: build/vessel/prod
build/vessel/prod:
	$(GO) build -ldflags "-w -s" -o build/vessel ./cmd/vessel

.PHONY: build/vessel/small
build/vessel/small: build/vessel/prod
	$(UPX) --lzma build/vessel

.PHONY: build/vessel/dev
build/vessel/dev:
	$(GO) build -tags debug -o build/vessel ./cmd/vessel

.PHONY: build/client/prod
build/client/prod:
	$(GO) build -ldflags "-w -s" -o build/client ./cmd/client

.PHONY: build/client/dev
build/client/dev:
	$(GO) build -tags debug -o build/client ./cmd/client

.PHONY: build/server/prod
build/server/prod:
	$(GO) build -ldflags "-w -s" -o build/server ./cmd/server

.PHONY: build/server/dev
build/server/dev:
	$(GO) build -tags debug -o build/server ./cmd/server

.PHONY: generate
generate:
	$(GO) generate ./...

.PHONY: run/vessel
run/vessel:
	$(GO) run -race -tags debug ./cmd/vessel

.PHONY: run/server
run/server:
	$(GO) run -race ./cmd/server start

.PHONY: test
test:
	$(GO) test -race ./...

.PHONY: docker/dev
docker/dev:
	docker-compose -f docker/docker-compose.yaml up --force-recreate

.PHONY: docs/dev
docs/dev:
	docsify serve docs/

.PHONY: lint
lint:
	$(LINTER) run
