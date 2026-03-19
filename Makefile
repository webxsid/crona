PROJECT_NAME := crona
PROJECT_VERSION := 0.2.0
PROJECT_DESCRIPTION := Local-first work kernel, TUI, and shared contracts
GO ?= go
GOCACHE ?= /tmp/crona-go-cache

.PHONY: help meta build test fmt vet lint install-lint run-kernel run-tui install-kernel install-tui seed-dev clear-dev release

help:
	@printf "%s %s\n" "$(PROJECT_NAME)" "$(PROJECT_VERSION)"
	@printf "%s\n" "$(PROJECT_DESCRIPTION)"
	@printf "\nTargets:\n"
	@printf "  make build           Build shared, kernel, tui, and cli\n"
	@printf "  make test            Run kernel tests\n"
	@printf "  make fmt             Format the Go workspace\n"
	@printf "  make vet             Vet the Go workspace\n"
	@printf "  make lint            Run golangci-lint with repo config\n"
	@printf "  make install-lint    Install golangci-lint into GOPATH/bin\n"
	@printf "  make run-kernel      Run the kernel daemon\n"
	@printf "  make run-tui         Run the terminal UI\n"
	@printf "  make install-kernel  Install crona-kernel into GOPATH/bin\n"
	@printf "  make install-tui     Build and install the TUI binary into ./bin\n"
	@printf "  make seed-dev        Seed dev data through the kernel\n"
	@printf "  make clear-dev       Clear dev data through the kernel\n"
	@printf "  make release VERSION=<tag>  Build release binaries and installer\n"
	@printf "  make meta            Print project metadata\n"

meta:
	@printf "name=%s\nversion=%s\ndescription=%s\n" "$(PROJECT_NAME)" "$(PROJECT_VERSION)" "$(PROJECT_DESCRIPTION)"

build:
	cd shared && GOCACHE=$(GOCACHE) $(GO) build ./...
	cd kernel && GOCACHE=$(GOCACHE) $(GO) build ./...
	cd tui && GOCACHE=$(GOCACHE) $(GO) build ./...
	cd cli && GOCACHE=$(GOCACHE) $(GO) build ./...

test:
	cd kernel && GOCACHE=$(GOCACHE) $(GO) test ./...

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
	cd tui && GOCACHE=$(GOCACHE) $(GO) run .

install-kernel:
	cd kernel && GOCACHE=$(GOCACHE) $(GO) install ./cmd/crona-kernel

install-tui:
	mkdir -p bin
	cd tui && GOCACHE=$(GOCACHE) $(GO) build -o ../bin/crona-tui .

seed-dev:
	sh ./scripts/dev_seed.sh

clear-dev:
	sh ./scripts/dev_clear.sh

release:
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required, e.g. make release VERSION=v0.2.0-beta.1"; exit 1; fi
	sh ./scripts/build_release.sh "$(VERSION)"
