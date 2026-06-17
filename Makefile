# 交叉编译：一次构建 macOS / Linux 各架构可执行文件
APP      := monitor
CMD      := ./cmd/monitor
BIN_DIR  := bin
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BUILD_DATE ?= $(shell date -u +%Y-%m-%d)
REPO     ?= https://github.com/Ben-unbelieveable/process-monitor

LDFLAGS  := -s -w \
	-X github.com/liubo/process-monitor/internal/version.Version=$(VERSION) \
	-X github.com/liubo/process-monitor/internal/version.BuildDate=$(BUILD_DATE) \
	-X github.com/liubo/process-monitor/internal/version.Repository=$(REPO)

PLATFORMS := \
	darwin/arm64 \
	darwin/amd64 \
	linux/arm64 \
	linux/amd64

.PHONY: all build local clean

all: build

build:
	@mkdir -p $(BIN_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		out=$(BIN_DIR)/$(APP)-$$os-$$arch; \
		echo "==> building $$out ($$os/$$arch) $(VERSION)"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $$out $(CMD); \
	done
	@echo "done: $(BIN_DIR)/$(APP)-{darwin,linux}-{arm64,amd64}"

local:
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP) $(CMD)

clean:
	rm -rf $(BIN_DIR)
