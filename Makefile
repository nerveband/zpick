VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

.PHONY: build test install clean

build:
	go build $(LDFLAGS) -o zpick ./cmd/zpick

test:
	go test ./... -v

install: build
	cp zpick $(HOME)/.local/bin/

clean:
	rm -f zpick

vet:
	go vet ./...
