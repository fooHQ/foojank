OUTPUT := vessel

export CGO_ENABLED=0

.PHONY: build/prod
build/prod:
	go build -ldflags "-w -s" -o build/${OUTPUT} ./cmd/vessel

.PHONY: build/small
build/small: build/prod
	upx --lzma build/${OUTPUT}

.PHONY: build/dev
build/dev:
	go build -tags debug -o build/${OUTPUT} ./cmd/vessel

.PHONY: generate
generate:
	go generate ./...

.PHONY: run/vessel
run/vessel:
	CGO_ENABLED=1 go run -race -tags debug ./cmd/vessel

.PHONY: run/client
run/client:
	CGO_ENABLED=1 go run -race -tags debug ./cmd/client
