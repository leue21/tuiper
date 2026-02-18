APP := tuiper
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP)
CONFIG ?= $(HOME)/.config/tuiper/config.json
GOCACHE ?= /tmp/go-build
GOENV := GOCACHE=$(GOCACHE)
ifdef GOMODCACHE
GOENV := $(GOENV) GOMODCACHE=$(GOMODCACHE)
endif
GO := $(GOENV) go

.PHONY: build run tidy fmt test check clean
.PHONY: help man

build:
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN) .

run:
	$(GO) run . -config $(CONFIG)

tidy:
	$(GO) mod tidy

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

check: fmt test build

help:
	$(GO) run . -h

man:
	MANPAGER=$${MANPAGER:-cat} man ./docs/tuiper.1

clean:
	rm -rf $(BIN_DIR)
