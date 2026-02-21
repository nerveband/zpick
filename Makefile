VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

.PHONY: build test install clean

build:
	go build $(LDFLAGS) -o zmosh-picker ./cmd/zmosh-picker

test:
	go test ./... -v

install: build
	cp zmosh-picker $(HOME)/.local/bin/

clean:
	rm -f zmosh-picker

vet:
	go vet ./...
