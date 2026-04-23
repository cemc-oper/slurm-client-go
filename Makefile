BINARY_NAME := slclient
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo unknown)

LDFLAGS := -ldflags "-s -w -X github.com/cemc-oper/slurm-client-go/cmd.version=$(VERSION) -X github.com/cemc-oper/slurm-client-go/cmd.commit=$(COMMIT) -X github.com/cemc-oper/slurm-client-go/cmd.buildDate=$(BUILD_DATE)"

BUILD_DIR := bin
CGO_ENABLED ?= 0
export CGO_ENABLED

PLATFORMS := linux/amd64 linux/arm64 windows/amd64 windows/arm64

.PHONY: all build clean build-all $(PLATFORMS)

all: build

build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) main.go

$(PLATFORMS):
	@mkdir -p $(BUILD_DIR)
	@GOOS=$(word 1,$(subst /, ,$@)) GOARCH=$(word 2,$(subst /, ,$@)) \
		go build $(LDFLAGS) \
		-o $(BUILD_DIR)/$(BINARY_NAME)-$(word 1,$(subst /, ,$@))-$(word 2,$(subst /, ,$@))$(if $(findstring windows,$(word 1,$(subst /, ,$@))),.exe,) \
		main.go

build-all: $(PLATFORMS)

clean:
	$(RM) -r $(BUILD_DIR)
