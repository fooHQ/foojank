OUTPUT := vessel

export CGO_ENABLED=0

.PHONY: build
build:
	go build -ldflags "-w -s" -o build/${OUTPUT} ./cmd/vessel

.PHONY: shrink
shrink: build
	upx --lzma ${OUTPUT}
