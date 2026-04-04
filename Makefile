PROJECT_NAME := crona
PROJECT_REPO := webxsid/crona
PROJECT_VERSION := 1.0.0-beta.1
PROJECT_DESCRIPTION := Local-first work kernel, TUI, and shared contracts
GO ?= go
GOCACHE ?= /tmp/crona-go-cache

ifeq ($(CRONA_ENV),Dev)
BIN_SUFFIX := -dev
else
BIN_SUFFIX :=
endif

CLI_BINARY := $(PROJECT_NAME)$(BIN_SUFFIX)
KERNEL_BINARY := $(PROJECT_NAME)-kernel$(BIN_SUFFIX)
TUI_BINARY := $(PROJECT_NAME)-tui$(BIN_SUFFIX)

.PHONY: help meta build test test-shared test-kernel test-tui test-cli fmt vet lint install-lint run-kernel run-tui install-kernel install-tui install-cli seed-dev clear-dev release

help:
	@printf "%s %s\n" "$(PROJECT_NAME)" "$(PROJECT_VERSION)"
	@printf "%s\n" "$(PROJECT_DESCRIPTION)"
	@printf "\nTargets:\n"
	@printf "  make build           Build shared, kernel, tui, and cli\n"
	@printf "  make test            Run shared, kernel, tui, and cli tests\n"
	@printf "  make test-shared     Run shared tests\n"
	@printf "  make test-kernel     Run kernel tests\n"
	@printf "  make test-tui        Run tui tests\n"
	@printf "  make test-cli        Run cli tests\n"
	@printf "  make fmt             Format the Go workspace\n"
	@printf "  make vet             Vet the Go workspace\n"
	@printf "  make lint            Run golangci-lint with repo config\n"
	@printf "  make install-lint    Install golangci-lint into GOPATH/bin\n"
	@printf "  make run-kernel      Run the kernel daemon\n"
	@printf "  make run-tui         Run the terminal UI\n"
	@printf "  make install-kernel  Build %s into ./bin\n" "$(KERNEL_BINARY)"
	@printf "  make install-tui     Build %s into ./bin\n" "$(TUI_BINARY)"
	@printf "  make install-cli     Build %s into ./bin\n" "$(CLI_BINARY)"
	@printf "  make seed-dev        Seed dev data through the kernel\n"
	@printf "  make clear-dev       Clear dev data through the kernel\n"
	@printf "  make release VERSION=<tag>  Build release binaries and installer\n"
	@printf "  make meta            Print project metadata\n"

meta:
	@printf "name=%s\nrepo=%s\nversion=%s\ndescription=%s\n" "$(PROJECT_NAME)" "$(PROJECT_REPO)" "$(PROJECT_VERSION)" "$(PROJECT_DESCRIPTION)"

build:
	cd shared && GOCACHE=$(GOCACHE) $(GO) build ./...
	cd kernel && GOCACHE=$(GOCACHE) $(GO) build ./...
	cd tui && GOCACHE=$(GOCACHE) $(GO) build ./...
	cd cli && GOCACHE=$(GOCACHE) $(GO) build ./...

test:
	cd shared && GOCACHE=$(GOCACHE) $(GO) test ./...
	cd kernel && GOCACHE=$(GOCACHE) $(GO) test ./...
	cd tui && GOCACHE=$(GOCACHE) $(GO) test ./...
	cd cli && GOCACHE=$(GOCACHE) $(GO) test ./...

test-shared:
	cd shared && GOCACHE=$(GOCACHE) $(GO) test ./...

test-kernel:
	cd kernel && GOCACHE=$(GOCACHE) $(GO) test ./...

test-tui:
	cd tui && GOCACHE=$(GOCACHE) $(GO) test ./...

test-cli:
	cd cli && GOCACHE=$(GOCACHE) $(GO) test ./...

fmt:
	GOCACHE=$(GOCACHE) $(GO) fmt ./...

vet:
	cd shared && GOCACHE=$(GOCACHE) $(GO) vet ./...
	cd kernel && GOCACHE=$(GOCACHE) $(GO) vet ./...
	cd tui && GOCACHE=$(GOCACHE) $(GO) vet ./...
	cd cli && GOCACHE=$(GOCACHE) $(GO) vet ./...

lint:
	sh ./scripts/lint.sh

install-lint:
	GOCACHE=$(GOCACHE) $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8

run-kernel:
	cd kernel && GOCACHE=$(GOCACHE) $(GO) run ./cmd/crona-kernel

run-tui:
	cd tui && PATH="$(CURDIR)/bin:$$PATH" GOCACHE=$(GOCACHE) $(GO) run .

install-kernel:
	mkdir -p bin
	cd kernel && GOCACHE=$(GOCACHE) $(GO) build -o ../bin/$(KERNEL_BINARY) ./cmd/crona-kernel

install-tui:
	mkdir -p bin
	cd tui && GOCACHE=$(GOCACHE) $(GO) build -o ../bin/$(TUI_BINARY) .

install-cli:
	mkdir -p bin
	cd cli && GOCACHE=$(GOCACHE) $(GO) build -o ../bin/$(CLI_BINARY) ./cmd/crona

seed-dev:
	sh ./scripts/dev_seed.sh

clear-dev:
	sh ./scripts/dev_clear.sh

release:
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required, e.g. make release VERSION=v1.0.0-beta.1"; exit 1; fi
	sh ./scripts/build_release.sh "$(VERSION)"
