PROJECT_NAME := crona
PROJECT_REPO := webxsid/crona
PROJECT_VERSION := 1.6.0-beta.4
PROJECT_DESCRIPTION := Local-first work kernel, TUI, and shared contracts
GO ?= go
GOCACHE ?= /tmp/crona-go-cache

ifeq ($(CRONA_ENV),Dev)
BIN_SUFFIX := -dev
else
BIN_SUFFIX :=
endif

CLI_BINARY := $(PROJECT_NAME)$(BIN_SUFFIX)
DAEMON_BINARY := $(PROJECT_NAME)-daemon$(BIN_SUFFIX)
TUI_BINARY := $(PROJECT_NAME)-tui$(BIN_SUFFIX)

.PHONY: help meta build test test-unit test-e2e test-coverage test-shared test-daemon test-tui test-cli fmt vet lint ci release-check install-lint install-fmt run-daemon run-tui install-daemon install-tui install-cli seed-dev clear-dev release brew-test brew-generate brew-clean brew-upgrade-test

help:
	@printf "%s %s\n" "$(PROJECT_NAME)" "$(PROJECT_VERSION)"
	@printf "%s\n" "$(PROJECT_DESCRIPTION)"
	@printf "\nTargets:\n"
	@printf "  make build           Build shared, daemon, tui, and cli\n"
	@printf "  make test            Run shared, daemon, tui, and cli tests\n"
	@printf "  make test-unit       Run non-e2e module tests\n"
	@printf "  make test-e2e        Run daemon IPC e2e tests\n"
	@printf "  make test-coverage   Generate module coverage summaries\n"
	@printf "  make test-shared     Run shared tests\n"
	@printf "  make test-daemon     Run daemon tests\n"
	@printf "  make test-tui        Run tui tests\n"
	@printf "  make test-cli        Run cli tests\n"
	@printf "  make fmt             Format the Go workspace with gofmt and golines\n"
	@printf "  make vet             Vet the Go workspace\n"
	@printf "  make lint            Run golangci-lint with repo config\n"
	@printf "  make ci              Run release metadata, tests, vet, lint, and coverage\n"
	@printf "  make release-check   Validate version and prerelease metadata consistency\n"
	@printf "  make install-lint    Install golangci-lint into GOPATH/bin\n"
	@printf "  make install-fmt     Install golines into GOPATH/bin\n"
	@printf "  make run-daemon      Run the daemon\n"
	@printf "  make run-tui         Run the terminal UI\n"
	@printf "  make install-daemon  Build %s into ./bin\n" "$(DAEMON_BINARY)"
	@printf "  make install-tui     Build %s into ./bin\n" "$(TUI_BINARY)"
	@printf "  make install-cli     Build %s into ./bin\n" "$(CLI_BINARY)"
	@printf "  make seed-dev        Seed dev data through the daemon\n"
	@printf "  make clear-dev       Clear dev data through the daemon\n"
	@printf "  make brew-test       Run isolated Homebrew validation against dist/\n"
	@printf "  make brew-generate   Generate isolated Homebrew tap and formula only\n"
	@printf "  make brew-upgrade-test  Simulate isolated Homebrew upgrade flow\n"
	@printf "  make brew-clean      Remove isolated Homebrew test artifacts\n"
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
	$(MAKE) test-unit

test-unit:
	cd shared && GOCACHE=$(GOCACHE) $(GO) test ./...
	cd kernel && GOCACHE=$(GOCACHE) $(GO) test ./...
	cd tui && GOCACHE=$(GOCACHE) $(GO) test ./...
	cd cli && GOCACHE=$(GOCACHE) $(GO) test ./...

test-e2e:
	cd kernel && GOCACHE=$(GOCACHE) $(GO) test -tags=e2e ./e2e

test-coverage:
	sh ./scripts/coverage.sh

test-shared:
	cd shared && GOCACHE=$(GOCACHE) $(GO) test ./...

test-daemon:
	cd kernel && GOCACHE=$(GOCACHE) $(GO) test ./...

test-tui:
	cd tui && GOCACHE=$(GOCACHE) $(GO) test ./...

test-cli:
	cd cli && GOCACHE=$(GOCACHE) $(GO) test ./...

fmt:
	sh ./scripts/fmt.sh

vet:
	cd shared && GOCACHE=$(GOCACHE) $(GO) vet ./...
	cd kernel && GOCACHE=$(GOCACHE) $(GO) vet ./...
	cd tui && GOCACHE=$(GOCACHE) $(GO) vet ./...
	cd cli && GOCACHE=$(GOCACHE) $(GO) vet ./...

lint:
	sh ./scripts/lint.sh

ci:
	$(MAKE) release-check
	$(MAKE) test-unit
	$(MAKE) vet
	$(MAKE) lint
	$(MAKE) test-coverage

release-check:
	sh ./scripts/check_release_metadata.sh

install-lint:
	GOCACHE=$(GOCACHE) $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4

install-fmt:
	GOCACHE=$(GOCACHE) $(GO) install github.com/segmentio/golines@v0.13.0

run-daemon:
	cd kernel && GOCACHE=$(GOCACHE) $(GO) run ./cmd/crona-kernel

run-tui:
	cd tui && PATH="$(CURDIR)/bin:$$PATH" GOCACHE=$(GOCACHE) $(GO) run .

install-daemon:
	mkdir -p bin
	cd kernel && GOCACHE=$(GOCACHE) $(GO) build -o ../bin/$(DAEMON_BINARY) ./cmd/crona-kernel

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
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required, e.g. make release VERSION=v1.0.0"; exit 1; fi
	sh ./scripts/build_release.sh "$(VERSION)"

brew-test:
	sh ./scripts/test_homebrew.sh test

brew-generate:
	sh ./scripts/test_homebrew.sh generate-only

brew-upgrade-test:
	sh ./scripts/test_homebrew.sh upgrade-test

brew-clean:
	sh ./scripts/test_homebrew.sh clean
