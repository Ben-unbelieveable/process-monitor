# 交叉编译：一次构建 macOS / Linux 各架构可执行文件
APP      := monitor
CMD      := ./cmd/monitor
BIN_DIR  := bin
LDFLAGS  := -s -w

# 目标平台：darwin arm64/amd64，linux arm64/amd64
PLATFORMS := \
	darwin/arm64 \
	darwin/amd64 \
	linux/arm64 \
	linux/amd64

.PHONY: all build clean

# 默认：交叉编译全部平台
all: build

build:
	@mkdir -p $(BIN_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		out=$(BIN_DIR)/$(APP)-$$os-$$arch; \
		echo "==> building $$out ($$os/$$arch)"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $$out $(CMD); \
	done
	@echo "done: $(BIN_DIR)/$(APP)-{darwin,linux}-{arm64,amd64}"

clean:
	rm -rf $(BIN_DIR)
