# Nebulant cli Makefile. 
# github.com/develation/nebulant-cli

VERSION = 0
PATCHLEVEL = 2
SUBLEVEL = 0
EXTRAVERSION = -beta
NAME =

######

CLIVERSION = $(VERSION)$(if $(PATCHLEVEL),.$(PATCHLEVEL)$(if $(SUBLEVEL),.$(SUBLEVEL)))$(EXTRAVERSION)
DATE = `date +'%Y%m%d'`

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
	GO111MODULE=on CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant$(GOEXE) nebulant.go

builddebug:
	GO111MODULE=on CGO_ENABLED=0 go build -a -trimpath -ldflags "$(LDFLAGS)" -o bin/nebulant-debug nebulant.go

.PHONY: buildlocal
buildlocal:
	go build -ldflags "$(LDFLAGS) $(LOCALLDFLAGS)" -trimpath -o bin/nebulant-dev-NOPROD nebulant.go

.PHONY: builddev
builddev:
	go build -ldflags "$(LDFLAGS) $(DEVLDFLAGS)" -trimpath -o bin/nebulant-dev-NOPROD nebulant.go

.PHONY: buildall
buildall:
	@echo "Building..."
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-linux-arm nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-linux-arm64 nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-linux-386 nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-linux-amd64 nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=freebsd GOARCH=386 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-freebsd-386 nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-freebsd-amd64 nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=freebsd GOARCH=arm64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-freebsd-arm64 nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=freebsd GOARCH=arm go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-freebsd-arm nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=openbsd GOARCH=386 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-openbsd-386 nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-openbsd-amd64 nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=openbsd GOARCH=arm64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-openbsd-arm64 nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=openbsd GOARCH=arm go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-openbsd-arm nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-windows-386.exe nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-windows-amd64.exe nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-windows-arm64.exe nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=windows GOARCH=arm go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-windows-arm.exe nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-darwin-amd64 nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-darwin-arm64 nebulant.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=js GOARCH=wasm go build -a -trimpath -ldflags "-w -s $(LDFLAGS)" -o bin/nebulant-$(CLIVERSION)-$(DATE)-js-wasm nebulant.go
	@echo "Check bin/ for builds"

.PHONY: prepare_reproducible_buildall
prepare_reproducible_buildall:
	@echo "This build needs go version $(GOVERSION) to be reproducible"
	go version | grep $(GOVERSION) || exit 1
	@echo "WARNING: This will override /tmp/nebulant-cli. Build will start after 10s..." && sleep 5
	@ echo "5..." && sleep 1
	@ echo "4..." && sleep 1
	@ echo "3..." && sleep 1
	@ echo "2..." && sleep 1
	@ echo "1..." && sleep 1
	ln -snf $(CLIPATH) /tmp/nebulant-cli && cd /tmp/nebulant-cli
	cd /tmp/nebulant-cli

.PHONY: reproducible_buildall
reproducible_buildall: prepare_reproducible_buildall buildall shasum

.PHONY: secure
secure:
	# https://github.com/securego/gosec/blob/master/README.md
	# G307 -- https://github.com/securego/gosec/issues/512
	$(GOPATH)/bin/gosec -exclude=G307 ./...
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