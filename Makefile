BINARY_NAME := multichat
BINARY_PATH := ./cmd/
PKG := github.com/llan0/multichat

# Version info
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# linker flags to inject version info
LDFLAGS := -ldflags "-X $(PKG)/internal/version.Version=$(VERSION) \
	-X $(PKG)/internal/version.Commit=$(COMMIT) \
	-X $(PKG)/internal/version.BuildTime=$(BUILD_TIME)"

.PHONY: all
all: build

.PHONY: build
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) $(BINARY_PATH)

.PHONY: run
run:
	go run $(LDFLAGS) $(BINARY_PATH)

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -f $(BINARY_NAME)

.PHONY: version
version:
	@echo $(VERSION)
