# 基本项目变量
# 应用名称
APP_NAME := hipush
# 本地可执行文件存放路径
LOCALBIN := $(shell pwd )/bin

# 镜像构建相关变量
REGISTRY := hipush.com
PROJECT := library
TAG ?= latest
IMG := $(REGISTRY)/$(PROJECT)/$(APP_NAME):$(TAG)

# 构建相关变量
VERSION ?= v1.0.0
BUILD_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
BUILD_COMMIT := $(shell git rev-parse HEAD)
BUILD_TIME := $(shell date '+%Y-%m-%d %H:%M:%S')
BUILD_GO_VERSION := $(shell go version | grep -o 'go[0-9].[0-9].*')
VERSION_PATH := github.com/cossteam/hipush/pkg/version
PLATFORM := linux/arm64,linux/amd64

# 编译标志
GOFLAGS := $(EXTRA_GOFLAGS)
GOGCFLAGS := ""
GOLDFLAGS := -X '${VERSION_PATH}.GitTag=${VERSION}' \
	-X '${VERSION_PATH}.GitBranch=${BUILD_BRANCH}' \
	-X '${VERSION_PATH}.GitCommit=${BUILD_COMMIT}' \
	-X '${VERSION_PATH}.BuildTime=${BUILD_TIME}' \
	-X '${VERSION_PATH}.GoVersion=${BUILD_GO_VERSION}' \
	-extldflags \"-static\"

# 调试标志
ifeq ($(DEBUG),1)
	# 调试模式 - 禁用优化和内联
	GOGCFLAGS := "all=-N -l"
else
	# 发布模式 - 禁用符号和 DWARF 信息，修剪嵌入的路径
	GOLDFLAGS := -s -w $(GOLDFLAGS)
	GOFLAGS += -trimpath
endif


.PHONY: all
all: build

$(LOCALBIN):
	@mkdir -p $(LOCALBIN)

.PHONY: build
## 编译项目并生成可执行文件
build: dep fmt vet
	echo $(PLATFORM) | tr ',' '\n' | while read plat; do \
		OS=$$(echo $$plat | cut -d'/' -f1); \
		ARCH=$$(echo $$plat | cut -d'/' -f2); \
		GOOS=$$OS GOARCH=$$ARCH CGO_ENABLED=0 go build -gcflags $(GOGCFLAGS) -ldflags "$(GOLDFLAGS)" -a -o $(LOCALBIN)/$(APP_NAME)-$$OS-$$ARCH $(GOFLAGS) ./cmd/hipush; \
	done

.PHONY: dep
## 获取依赖项
dep:
	@go mod tidy

.PHONY: run
## 运行开发服务器
run: dep
	@go run ./cmd/hipush/

.PHONY: vet
## 运行 go vet 检查代码
vet:
	go vet ./...

.PHONY: test
## 运行所有测试
test: fmt vet
	go test -v -short -race -covermode=atomic ./...

.PHONY: lint
## 运行所有文件进行对应的静态检查任务
lint: lint-go lint-markdown lint-yaml

.PHONY: fmt
## 格式化对应的文件内容
fmt: fmt-go

.PHONY: fmt-go
## 格式化 Go 代码
fmt-go:
	go fmt ./...


.PHONY: install
## 下载需要用到的工具
install: install-swag install-golangci-lint install-gofumptn install-license-checker

GO_SWAGGER_VERSION = v1.16.4
.PHONY: install-swag
install-swag:
	if ! test -x $(LOCALBIN)/swag || ! $(LOCALBIN)/swag --version | grep $(GO_SWAGGER_VERSION) >/dev/null; then \
		GOBIN=$(LOCALBIN) go install github.com/swaggo/swag/cmd/swag@$(GO_SWAGGER_VERSION); \
	fi

GOLANGCI_LINT_VERSION = v1.62.2
.PHONY: install-golangci-lint
install-golangci-lint:
	if ! test -x $(LOCALBIN)/golangci-lint || ! $(LOCALBIN)/golangci-lint --version | grep $(GOLANGCI_LINT_VERSION) >/dev/null; then \
		GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi

LICENSE_CHECKER_VERSION = v0.6.0
.PHONY: install-license-checker
install-license-checker: $(LOCALBIN)
	if ! test -x $(LOCALBIN)/license-eye || ! $(LOCALBIN)/license-eye --version | grep $(LICENSE_CHECKER_VERSION) >/dev/null; then \
		GOBIN=$(LOCALBIN) go install github.com/apache/skywalking-eyes/cmd/license-eye@$(LICENSE_CHECKER_VERSION); \
	fi

OAPI_CODEGEN_VERSION = v2.4.1
.PHONY: install-oapi-codegen
install-oapi-codegen: $(LOCALBIN)
	if ! test -x $(LOCALBIN)/oapi-codegen || ! $(LOCALBIN)/oapi-codegen --version | grep $(OAPI_CODEGEN_VERSION) >/dev/null; then \
		GOBIN=$(LOCALBIN) go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION); \
	fi

GOFUMPTN_VERSION = v0.7.0
.PHONY: install-gofumptn
install-gofumptn: $(LOCALBIN)
	if ! test -x $(LOCALBIN)/gofumpt || ! $(LOCALBIN)/gofumpt --version | grep $(GOFUMPTN_VERSION) >/dev/null; then \
		GOBIN=$(LOCALBIN) go install mvdan.cc/gofumpt@$(GOFUMPTN_VERSION); \
	fi

.PHONY: lint-go
## 对 Go 代码进行静态分析
lint-go: install-golangci-lint install-gofumptn
	$(LOCALBIN)/golangci-lint run
	$(LOCALBIN)/gofumpt -l -w .

.PHONY: lint-markdown
## 对 Markdown 文件进行检查
lint-markdown:
	if ! command -v markdownlint >/dev/null 2>&1; then npm install -g markdownlint-cli; fi
	@# ignore markdown under 'xxx/xxx'
	- markdownlint '{*.md,site/**/*.md}' --disable MD012 MD013 MD029 MD033 MD034 MD036 MD041

# 我们不使用 if ! command -v yamllint 因为某些环境可能已经预安装了 Python 版本。
# 检查特定路径确保我们使用的是 Node.js 版本，以避免冲突。
.PHONY: lint-yaml
## 对 YAML 文件进行静态分析
lint-yaml:
	YAML_LINT="$$(npm config get prefix)/bin/yamllint"; \
	if ! test -x "$${YAML_LINT}"; then \
		npm install -g yaml-lint; \
	fi; \
	"$${YAML_LINT}" "**/*.(yaml|yml)"


.PHONY: gen
## 自动生成代码(proto、openapi...)
gen: gen-openapi
	go generate ./...

.PHONY: gen-openapi
## 生成 OpenAPI 规范
gen-openapi:
	# swag init --generalInfo cmd/main.go --output ./api/openapi/v2
	@echo "OpenAPI spec generated successfully in ./api/openapi/v2/swagger.json"

	@$(LOCALBIN)/oapi-codegen -generate gin,spec -o api/http/v1/openapi.gen.go -package v1 api/openapi/v3/*.yaml
	@$(LOCALBIN)/oapi-codegen -generate types -o api/http/v1/openapi.types.go -package v1 api/openapi/v3/*.yaml


.PHONY: clean
## 清理生成的文件
clean:
	rm -rf $(LOCALBIN)


.PHONY: docker-build
## 构建 Docker 镜像
docker-build:
	docker build --platform=$(PLATFORM)  \
	--build-arg BUILD_BRANCH="$(BUILD_BRANCH)" \
	--build-arg BUILD_COMMIT="$(BUILD_COMMIT)" \
	--build-arg BUILD_TIME="$(BUILD_TIME)" \
	--build-arg GOGCFLAGS="$(GOGCFLAGS)" \
	--build-arg GOLDFLAGS="$(GOLDFLAGS)" \
	--build-arg GOFLAGS="$(GOFLAGS)" \
	--build-arg APP_NAME="$(APP_NAME)" \
	-t $(IMG) .

HOST_PORT ?= 8080
CONTAINER_PORT ?= 8080
.PHONY: docker-run
## 启动 Docker 容器
docker-run:
	docker run -d -p $(HOST_PORT):$(CONTAINER_PORT) $(IMG)

.PHONY: docker-push
## 推送 Docker 镜像到远程仓库
docker-push:
	docker push $(IMG)


.PHONY: help
## 显示 Makefile 帮助信息
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)
