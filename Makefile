# gorp Makefile

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags="-s -w -X github.com/ngq/gorp/cmd/gorp/cmd.Version=$(VERSION) -X github.com/ngq/gorp/cmd/gorp/cmd.Commit=$(COMMIT) -X github.com/ngq/gorp/cmd/gorp/cmd.BuildDate=$(BUILD_DATE)"

.PHONY: all
all: build

# 构建 CLI
.PHONY: build
build:
	cd cmd/gorp && go build $(LDFLAGS) -o gorp.exe .

# 安装 CLI 到 GOPATH/bin
.PHONY: install
install:
	cd cmd/gorp && go install $(LDFLAGS)

# 运行测试
.PHONY: test
test:
	go test ./... -cover

# 运行测试（带覆盖率报告）
.PHONY: test-coverage
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# 代码检查
.PHONY: lint
lint:
	golangci-lint run ./...

# 清理构建产物
.PHONY: clean
clean:
	rm -f cmd/gorp/gorp
	rm -f cmd/gorp/gorp.exe
	rm -f coverage.out coverage.html

# 格式化代码
.PHONY: fmt
fmt:
	go fmt ./...

# 更新依赖
.PHONY: tidy
tidy:
	go mod tidy
	cd cmd/gorp && go mod tidy

# 构建 all 子模块
.PHONY: build-all
build-all: build
	go build ./...
