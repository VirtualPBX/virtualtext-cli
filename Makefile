.PHONY: build test install

BINARY := $(CURDIR)/bin/vt

build:
	go build -o $(BINARY) ./cmd/vt

test:
	go test ./...

install: build
	install $(BINARY) $(HOME)/.local/bin/vt
