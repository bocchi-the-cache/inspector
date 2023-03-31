GO := go
GOOS := linux
GO_ENVS := CGO_ENABLED=0 GOOS=${GOOS} GOPROXY=direct GOSUMDB=off
GO_TOOL_ENVS := GO111MODULE=off CGO_ENABLED=0 GOPROXY=direct GOSUMDB=off

BUILD_TIME = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_SHA := $(shell git rev-parse HEAD 2>/dev/null)
VERSION := $(shell cat VERSION | tr -d '\n')
ENV := dev

GO_LD_FLAGS = \
	-X main.BuildTime=${BUILD_TIME} \
	-X main.GitSHA=${GIT_SHA} \
	-X main.Version=${VERSION} \
	-X main.Env=${ENV}

GO_BUILD_FLAGS = -v -ldflags='${GO_LD_FLAGS}'

BINARIES = inspector
VERSION_NAME = ${BINARIES}-${VERSION}
DIST := dist/${VERSION_NAME}
TAR = ${VERSION_NAME}.tar.gz
PKGS = $(shell ${GO} list ./... | tr '\n' ',')
EXEC := ${DIST}/${BINARIES}

.PHONY: all dep build clean pkg

all: clean pkg

build: dep
	env "PATH=${PATH}:$(shell go env GOPATH)/bin" ${GO_TOOL_ENVS} ${GO} generate ./...
	env ${GO_ENVS} ${GO} build ${GO_BUILD_FLAGS} -o ${EXEC} .
	test -f config/config.yaml
	mkdir -p ${DIST}/config/
	cp -f config/config.yaml ${DIST}/config/config.yaml

clean:
	rm -rf ${TAR}
	rm -rf dist

pkg: build
	chmod a+x ${EXEC}

	tar -zcf ${TAR} -C dist ${VERSION_NAME}
	@echo "Build successfully"


fmt:
	goimports -w .
