VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

.PHONY: build test install clean

build:
	go build $(LDFLAGS) -o zp ./cmd/zp

test:
	go test ./... -v

install: build
	cp zp $(HOME)/.local/bin/
ifeq ($(shell uname),Darwin)
	xattr -cr $(HOME)/.local/bin/zp
	codesign -fs - $(HOME)/.local/bin/zp
endif
	@ln -sf $(HOME)/.local/bin/zp /usr/local/bin/zp 2>/dev/null || \
		echo "  note: run 'sudo ln -sf $(HOME)/.local/bin/zp /usr/local/bin/zp' for system-wide PATH"

clean:
	rm -f zp zpick

vet:
	go vet ./...
