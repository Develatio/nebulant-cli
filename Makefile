# Nebulant cli Makefile. 
# github.com/develation/nebulant-cli

VERSION = 0
PATCHLEVEL = 2
SUBLEVEL = 0
EXTRAVERSION := -beta-git-$(shell git log -1 --format=%h)
NAME =

######

CLIVERSION = $(VERSION)$(if $(PATCHLEVEL),.$(PATCHLEVEL)$(if $(SUBLEVEL),.$(SUBLEVEL)))$(EXTRAVERSION)
DATE = $(shell git log -1 --date=format:'%Y%m%d' --format=%cd)

PRERELEASE = true
ifeq ($(shell expr $(PATCHLEVEL) % 2), 0)
	PRERELEASE = false
endif

PKG_LIST := $(shell go list ./... | grep -v /vendor/)



LDFLAGS = -X github.com/develatio/nebulant-cli/config.Version=$(CLIVERSION)\
	-X github.com/develatio/nebulant-cli/config.VersionDate=$(DATE)

LOCALLDFLAGS = -X github.com/develatio/nebulant-cli/config.WSScheme=ws\
	-X github.com/develatio/nebulant-cli/config.BackendProto=https\
	-X github.com/develatio/nebulant-cli/config.BackendURLDomain=api.nebulant.lc\
	-X github.com/develatio/nebulant-cli/config.AccountURLDomain=account.nebulant.lc\
	-X github.com/develatio/nebulant-cli/config.PanelURLDomain=panel.nebulant.lc\
	-X github.com/develatio/nebulant-cli/config.FrontUrl=https://builder.nebulant.lc\
	-X github.com/develatio/nebulant-cli/config.FrontOrigin=*\
	-X github.com/develatio/nebulant-cli/config.AssetDescriptorURL=https://builder-assets.nebulant.dev/assets.json

DEVLDFLAGS = -X github.com/develatio/nebulant-cli/config.WSScheme=wss\
	-X github.com/develatio/nebulant-cli/config.BackendProto=https\
	-X github.com/develatio/nebulant-cli/config.BackendURLDomain=api.nebulant.dev\
	-X github.com/develatio/nebulant-cli/config.FrontUrl=https://builder.nebulant.dev\
	-X github.com/develatio/nebulant-cli/config.FrontOrigin=*\
	-X github.com/develatio/nebulant-cli/config.AssetDescriptorURL=https://builder-assets.nebulant.dev/assets.json

MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
CLIPATH := $(realpath $(dir $(MAKEFILE_PATH)))
GOVERSION = 1.21.0

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

.PHONY: runrace
runrace:
	CGO_ENABLED=1 go run -race -ldflags "$(LDFLAGS) $(LOCALLDFLAGS)" nebulant.go $(ARGS)

.PHONY: run
run:
	go run -ldflags "$(LDFLAGS) $(LOCALLDFLAGS)" nebulant.go $(ARGS)

.PHONY: rundev
rundev:
	go run -race -ldflags "$(LDFLAGS) $(DEVLDFLAGS)" nebulant.go $(ARGS)

.PHONY: rundockerdev
rundockerdev:
	# ej: make rundockerdev ARGS="-x -s -b 0.0.0.0:15678"
	docker run --rm -v "$(PWD)":/usr/src/myapp -w /usr/src/myapp -p 15678:15678 golang:$(GOVERSION) go run -race -ldflags "$(LDFLAGS) $(DEVLDFLAGS)" nebulant.go $(ARGS)

.PHONY: build
build:
	mkdir -p dist/$(CLIVERSION)
	GO111MODULE=on CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/nebulant$(GOEXE) nebulant.go
	shasum dist/nebulant$(GOEXE) > dist/nebulant.shasum

builddebug:
	GO111MODULE=on CGO_ENABLED=0 go build -a -trimpath -ldflags "$(LDFLAGS)" -o dist/nebulant-debug nebulant.go

.PHONY: buildlocal
buildlocal:
	go build -ldflags "$(LDFLAGS) $(LOCALLDFLAGS)" -trimpath -o dist/nebulant-dev-NOPROD nebulant.go

.PHONY: builddev
builddev:
	go build -ldflags "$(LDFLAGS) $(DEVLDFLAGS)" -trimpath -o dist/nebulant-dev-NOPROD nebulant.go

.PHONY: buildall
buildall:
	@echo "Building..."
	mkdir -p dist/$(CLIVERSION)
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-linux-arm nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-linux-arm > dist/$(CLIVERSION)/nebulant-linux-arm.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-linux-arm64 nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-linux-arm64 > dist/$(CLIVERSION)/nebulant-linux-arm64.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-linux-386 nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-linux-386 > dist/$(CLIVERSION)/nebulant-linux-386.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-linux-amd64 nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-linux-amd64 > dist/$(CLIVERSION)/nebulant-linux-amd64.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=freebsd GOARCH=386 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-freebsd-386 nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-freebsd-386 > dist/$(CLIVERSION)/nebulant-freebsd-386.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-freebsd-amd64 nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-freebsd-amd64 > dist/$(CLIVERSION)/nebulant-freebsd-amd64.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=freebsd GOARCH=arm64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-freebsd-arm64 nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-freebsd-arm64 > dist/$(CLIVERSION)/nebulant-freebsd-arm64.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=freebsd GOARCH=arm go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-freebsd-arm nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-freebsd-arm > dist/$(CLIVERSION)/nebulant-freebsd-arm.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=openbsd GOARCH=386 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-openbsd-386 nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-openbsd-386 > dist/$(CLIVERSION)/nebulant-openbsd-386.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-openbsd-amd64 nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-openbsd-amd64 > dist/$(CLIVERSION)/nebulant-openbsd-amd64.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=openbsd GOARCH=arm64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-openbsd-arm64 nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-openbsd-arm64 > dist/$(CLIVERSION)/nebulant-openbsd-arm64.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=openbsd GOARCH=arm go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-openbsd-arm nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-openbsd-arm > dist/$(CLIVERSION)/nebulant-openbsd-arm.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-windows-386.exe nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-windows-386.exe > dist/$(CLIVERSION)/nebulant-windows-386.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-windows-amd64.exe nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-windows-amd64.exe > dist/$(CLIVERSION)/nebulant-windows-amd64.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-windows-arm64.exe nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-windows-arm64.exe > dist/$(CLIVERSION)/nebulant-windows-arm64.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=windows GOARCH=arm go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-windows-arm.exe nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-windows-arm.exe > dist/$(CLIVERSION)/nebulant-windows-arm.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-darwin-amd64 nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-darwin-amd64 > dist/$(CLIVERSION)/nebulant-darwin-amd64.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-darwin-arm64 nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-darwin-arm64 > dist/$(CLIVERSION)/nebulant-darwin-arm64.shasum
	GO111MODULE=on CGO_ENABLED=0 GOOS=js GOARCH=wasm go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o dist/$(CLIVERSION)/nebulant-js-wasm nebulant.go
	shasum dist/$(CLIVERSION)/nebulant-js-wasm > dist/$(CLIVERSION)/nebulant-js-wasm.shasum
	@echo "Check dist/ for builds"

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
	@echo $(GOVERSION)

.PHONY: shasum
shasum:
	shasum bin/*