VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

.PHONY: build test install clean

build:
	go build $(LDFLAGS) -o zp ./cmd/zp

test:
	go test ./... -v

install: build
	cp zp $(HOME)/.local/bin/

clean:
	rm -f zp zpick

vet:
	go vet ./...
