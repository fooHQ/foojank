export CGO_ENABLED=0

.PHONY: build/vessel/prod
build/vessel/prod:
	go build -ldflags "-w -s" -o build/vessel ./cmd/vessel

.PHONY: build/vessel/small
build/vessel/small: build/vessel/prod
	upx --lzma build/vessel

.PHONY: build/vessel/dev
build/vessel/dev:
	go build -tags debug -o build/vessel ./cmd/vessel

.PHONY: build/client/prod
build/client/prod:
	go build -ldflags "-w -s" -o build/client ./cmd/client

.PHONY: build/client/dev
build/client/dev:
	go build -tags debug -o build/client ./cmd/client

.PHONY: generate
generate:
	go generate ./...

.PHONY: run/vessel
run/vessel:
	CGO_ENABLED=1 go run -race -tags debug ./cmd/vessel

.PHONY: test
test:
	CGO_ENABLED=1 go test -race ./...

.PHONY: docker/dev
docker/dev:
	docker-compose -f docker/docker-compose.yaml up --force-recreate

.PHONY: docs/dev
docs/dev:
	docsify serve docs/
