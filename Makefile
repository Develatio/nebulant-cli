# Nebulant cli Makefile.
# github.com/develatio/nebulant-cli

VERSION = 0
PATCHLEVEL = 5
SUBLEVEL = 0
EXTRAVERSION = -beta
# EXTRAVERSION := -beta-git-$(shell git log -1 --format=%h)
NAME =

######

CLIVERSION = $(VERSION)$(if $(PATCHLEVEL),.$(PATCHLEVEL)$(if $(SUBLEVEL),.$(SUBLEVEL)))$(EXTRAVERSION)
DATE = $(shell git log -1 --date=format:'%Y%m%d' --format=%cd)
COMMIT = $(shell git log -1 --format=%h)
GOVERSION = $(shell go env GOVERSION)

PRERELEASE = true
ifeq ($(shell expr $(PATCHLEVEL) % 2), 0)
	PRERELEASE = false
endif

PKG_LIST := $(shell go list ./... | grep -v /vendor/)


LDFLAGS = -X github.com/develatio/nebulant-cli/config.Version=$(CLIVERSION)\
	-X github.com/develatio/nebulant-cli/config.VersionDate=$(DATE)\
	-X github.com/develatio/nebulant-cli/config.VersionCommit=$(COMMIT)\
	-X 'github.com/develatio/nebulant-cli/config.VersionGo=$(GOVERSION)'

LOCALLDFLAGS = -X github.com/develatio/nebulant-cli/config.WSScheme=ws\
	-X github.com/develatio/nebulant-cli/config.BASE_SCHEME=https\
	-X github.com/develatio/nebulant-cli/config.BACKEND_API_HOST=api.nebulant.lc\
	-X github.com/develatio/nebulant-cli/config.BACKEND_ACCOUNT_HOST=account.nebulant.lc\
	-X github.com/develatio/nebulant-cli/config.MARKET_API_HOST=market.nebulant.lc\
	-X github.com/develatio/nebulant-cli/config.PANEL_HOST=panel.nebulant.lc\
	-X github.com/develatio/nebulant-cli/config.FrontUrl=https://builder.nebulant.lc\
	-X github.com/develatio/nebulant-cli/config.UpdateDescriptorURL=https://releases.nebulant.lc/version.json\
	-X github.com/develatio/nebulant-cli/config.FrontOrigin=*\
	-X github.com/develatio/nebulant-cli/config.AssetDescriptorURL=https://builder-assets.nebulant.dev/assets.json

DEVLDFLAGS = -X github.com/develatio/nebulant-cli/config.WSScheme=wss\
	-X github.com/develatio/nebulant-cli/config.BASE_SCHEME=https\
	-X github.com/develatio/nebulant-cli/config.BACKEND_API_HOST=api.nebulant.dev\
	-X github.com/develatio/nebulant-cli/config.BACKEND_ACCOUNT_HOST=account.nebulant.dev\
	-X github.com/develatio/nebulant-cli/config.MARKET_API_HOST=market.nebulant.dev\
	-X github.com/develatio/nebulant-cli/config.PANEL_HOST=panel.nebulant.dev\
	-X github.com/develatio/nebulant-cli/config.FrontUrl=https://builder.nebulant.dev\
	-X github.com/develatio/nebulant-cli/config.UpdateDescriptorURL=https://releases.nebulant.dev/version.json\
	-X github.com/develatio/nebulant-cli/config.FrontOrigin=*\
	-X github.com/develatio/nebulant-cli/config.AssetDescriptorURL=https://builder-assets.nebulant.dev/assets.json

MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
CLIPATH := $(realpath $(dir $(MAKEFILE_PATH)))
MINGOVERSION = 1.21.0

ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

ifndef $(GOOS)
    GOOS=$(shell go env GOOS)
    export GOOS
endif

ifndef $(GOARCH)
    GOARCH=$(shell go env GOARCH)
    export GOARCH
endif

GOEXE=$(shell go env GOEXE)

.PHONY: create-network
create-network:
	docker network create nebulant-lan 2> /dev/null || true

.PHONY: runrace
runrace:
	CGO_ENABLED=1 go run -race -ldflags "-X github.com/develatio/nebulant-cli/config.LOAD_CONF_FILES=false $(LDFLAGS) $(LOCALLDFLAGS)" nebulant.go $(ARGS)

.PHONY: runracebridge
runracebridge:
	CGO_ENABLED=1 go run -race -ldflags "-X github.com/develatio/nebulant-cli/config.LOAD_CONF_FILES=false $(LDFLAGS) $(LOCALLDFLAGS)" ./bridge $(ARGS)

.PHONY: runbridge
runbridge:
	CGO_ENABLED=1 go run -ldflags "-X github.com/develatio/nebulant-cli/config.LOAD_CONF_FILES=false $(LDFLAGS) $(LOCALLDFLAGS)" ./bridge $(ARGS)

.PHONY: runbridgedocker
runbridgedocker: create-network
	docker compose -f docker-compose.yml up bridge

.PHONY: run
run:
	go run -ldflags "$(LDFLAGS) $(LOCALLDFLAGS)" nebulant.go $(ARGS)

.PHONY: rundev
rundev:
	go run -ldflags "$(LDFLAGS) $(DEVLDFLAGS)" nebulant.go $(ARGS)

.PHONY: rundockerdev
rundockerdev:
	# ej: make rundockerdev ARGS="-x serve -b 0.0.0.0:15678"
	docker run --rm -v "$(PWD)":/usr/src/myapp -w /usr/src/myapp -p 15678:15678 golang:$(MINGOVERSION) go run -race -ldflags "$(LDFLAGS) $(DEVLDFLAGS)" nebulant.go $(ARGS)

.PHONY: build
build:
	GO111MODULE=on CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/nebulant$(GOEXE) nebulant.go
	shasum dist/nebulant$(GOEXE) > dist/nebulant.checksum

buildbridge:
	GO111MODULE=on CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -a -trimpath -ldflags "-w -s -X github.com/develatio/nebulant-cli/config.LOAD_CONF_FILES=false $(LDFLAGS)" -o dist/nebulant-bridge$(GOEXE) ./bridge

builddebug:
	GO111MODULE=on CGO_ENABLED=0 go build -a -trimpath -ldflags "$(LDFLAGS)" -o dist/nebulant-debug nebulant.go

.PHONY: buildlocal
buildlocal:
	go build -ldflags "$(LDFLAGS) $(LOCALLDFLAGS)" -trimpath -o dist/nebulant-dev-NOPROD nebulant.go

.PHONY: builddev
builddev:
	go build -ldflags "$(LDFLAGS) $(DEVLDFLAGS)" -trimpath -o dist/nebulant-dev-NOPROD nebulant.go

.PHONY: build_platform
build_platform:
	@mkdir -p dist/v$(CLIVERSION)
	@GO111MODULE=on CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -a -trimpath -ldflags "-w -s $(LDFLAGS) $(EXTRAFLAGS)" -o dist/v$(CLIVERSION)/nebulant$(DIST_SUFFIX) nebulant.go
	@shasum -a 256 dist/v$(CLIVERSION)/nebulant$(DIST_SUFFIX) > dist/v$(CLIVERSION)/nebulant$(DIST_SUFFIX).checksum
	@printf "sha256: "
	@cat dist/v$(CLIVERSION)/nebulant$(DIST_SUFFIX).checksum

.PHONY: buildall
buildall:
	GOOS=linux GOARCH=arm GOEXE= DIST_SUFFIX=-linux-arm $(MAKE) build_platform
	GOOS=linux GOARCH=arm64 GOEXE= DIST_SUFFIX=-linux-arm64 $(MAKE) build_platform
	GOOS=linux GOARCH=386 GOEXE= DIST_SUFFIX=-linux-386 $(MAKE) build_platform
	GOOS=linux GOARCH=amd64 GOEXE= DIST_SUFFIX=-linux-amd64 $(MAKE) build_platform
	GOOS=freebsd GOARCH=arm GOEXE= DIST_SUFFIX=-freebsd-arm $(MAKE) build_platform
	GOOS=freebsd GOARCH=arm64 GOEXE= DIST_SUFFIX=-freebsd-arm64 $(MAKE) build_platform
	GOOS=freebsd GOARCH=386 GOEXE= DIST_SUFFIX=-freebsd-386 $(MAKE) build_platform
	GOOS=freebsd GOARCH=amd64 GOEXE= DIST_SUFFIX=-freebsd-amd64 $(MAKE) build_platform
	GOOS=openbsd GOARCH=arm GOEXE= DIST_SUFFIX=-openbsd-arm $(MAKE) build_platform
	GOOS=openbsd GOARCH=arm64 GOEXE= DIST_SUFFIX=-openbsd-arm64 $(MAKE) build_platform
	GOOS=openbsd GOARCH=386 GOEXE= DIST_SUFFIX=-openbsd-386 $(MAKE) build_platform
	GOOS=openbsd GOARCH=amd64 GOEXE= DIST_SUFFIX=-openbsd-amd64 $(MAKE) build_platform
	GOOS=windows GOARCH=arm GOEXE=.exe DIST_SUFFIX=-windows-arm.exe $(MAKE) build_platform
	GOOS=windows GOARCH=arm64 GOEXE=.exe DIST_SUFFIX=-windows-arm64.exe $(MAKE) build_platform
	GOOS=windows GOARCH=386 GOEXE=.exe DIST_SUFFIX=-windows-386.exe $(MAKE) build_platform
	GOOS=windows GOARCH=amd64 GOEXE=.exe DIST_SUFFIX=-windows-amd64.exe $(MAKE) build_platform
	GOOS=darwin GOARCH=arm64 GOEXE= DIST_SUFFIX=-darwin-arm64 $(MAKE) build_platform
	GOOS=darwin GOARCH=amd64 GOEXE= DIST_SUFFIX=-darwin-amd64 $(MAKE) build_platform
	# GOOS=js GOARCH=wasm GOEXE= DIST_SUFFIX=-js-wasm $(MAKE) build_platform

.PHONY: buildalldev
buildalldev: EXTRAFLAGS=$(DEVLDFLAGS)
buildalldev: buildall
	echo "done dev build"

.PHONY: secure
secure:
	# https://github.com/securego/gosec/blob/master/README.md
	# G307 -- https://github.com/securego/gosec/issues/512
	$(GOPATH)/bin/gosec -exclude=G307 ./...

.PHONY: staticanalysis
staticanalysis:
	# https://github.com/praetorian-inc/gokart
	$(GOPATH)/bin/gokart scan -v

.PHONY: unittest
unittest:
	go test -v -race $(PKG_LIST)

.PHONY: cover
cover:
	go test -cover -v -race $(PKG_LIST)

.PHONY: htmlcover
htmlcover:
	go test -coverprofile cover.out -v -race $(PKG_LIST)
	go tool cover -html=cover.out

.PHONY: cliversion
cliversion:
	@echo $(CLIVERSION)

.PHONY: ispre
ispre:
	@echo $(PRERELEASE)

.PHONY: versiondate
versiondate:
	@echo $(DATE)

.PHONY: goversion
goversion:
	@echo $(MINGOVERSION)