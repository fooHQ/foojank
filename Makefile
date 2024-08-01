OUTPUT := vessel

export CGO_ENABLED=0

.PHONY: build
build:
	go build -ldflags "-w -s" -o build/${OUTPUT} ./cmd/vessel

.PHONY: generate
generate:
	go generate ./...

.PHONY: shrink
shrink: build
	upx --lzma ${OUTPUT}

.PHONY: run
run:
	go run -tags debug ./cmd/vessel
