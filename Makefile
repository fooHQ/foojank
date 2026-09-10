BUILD_DIR = ./build
GO = devbox run go
GOLANGCI_LINT = devbox run golangci-lint

.PHONY: all
all: foojank foojankd

.PHONY: foojank
foojank:
	CGO_ENABLED=0 $(GO) build -o "$(BUILD_DIR)/$@" "./cmd/$@"

.PHONY: foojankd
foojankd:
	CGO_ENABLED=0 $(GO) build -o "$(BUILD_DIR)/$@" "./cmd/$@"

.PHONY: test
test:
	CGO_ENABLED=1 $(GO) test -race -timeout 30s -tags dev ./...

.PHONY: lint
lint:
	$(GOLANGCI_LINT) run --timeout 10m
