BIN_FILE=mix-scheduler-plugins
VERSION=latest
DEP_VERSION ?= v1.30.3
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOROOT ?= $(shell go env GOROOT)
VERSION_DIR="github.com/helen-frank/scansend/pkg/version"
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_TREE_STATE=$(shell if git status|grep -q 'clean';then echo clean; else echo dirty; fi)
GOVERSION=${shell go version}
# LDFLAGS="-s -w -X '${VERSION_DIR}.version=${VERSION}' -X '${VERSION_DIR}.gitCommit=${GIT_COMMIT}' -X '${VERSION_DIR}.gitTreeState=${GIT_TREE_STATE}' -X '${VERSION_DIR}.buildDate=${BUILD_DATE}' -X '${VERSION_DIR}.goVersion=${GOVERSION}'"
LDFLAGS="-s -w"

depUpdate:
	@rm -rf go.mod go.sum
	@go mod init github.com/helen-frank/mix-scheduler-plugins
	@bash hack/mod.sh ${DEP_VERSION}
	@go mod tidy
	@go mod vendor

build:
	@GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags ${LDFLAGS} -o _output/${GOOS}_${GOARCH}/${BIN_FILE} ./cmd/scheduler

dockerBuild:
	@docker build -t helenfrank/mix-scheduler-plugins:${VERSION} .

all:
	@make depUpdate
	@make build

cleanDir:
	@rm -rf _tmp/*
	@rm -rf _output/*

cleanBuild:
	@go clean

help:
	@echo "make; 更新依赖 格式化go代码 并编译生成二进制文件"
	@echo "make depUpdate; 更新依赖"
	@echo "make build; 编译go代码生成二进制文件"
	@echo "make test; 执行测试case"

.PHONY: build depUpdate test help cleanBuild cleanDir
