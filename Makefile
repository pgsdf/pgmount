# Makefile for pgmount

PREFIX ?= /usr/local
BINDIR = $(PREFIX)/bin
DATADIR = $(PREFIX)/share
MANDIR = $(PREFIX)/man

GO ?= go
INSTALL ?= install
MKDIR ?= mkdir -p
CHMOD ?= chmod

VERSION = 1.0.0

# Build flags
GOFLAGS = -ldflags "-X main.Version=$(VERSION)"

# Binaries
BINARIES = pgmountd pgmount pgumount pginfo

.PHONY: all build install uninstall clean test deps man

all: deps build

build: $(BINARIES)

deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	@echo "Dependencies ready."

pgmountd:
	$(GO) build $(GOFLAGS) -o $@ .

pgmount:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/pgmount

pgumount:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/pgumount

pginfo:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/pginfo

man:
	@echo "Generating man pages..."
	@if command -v pandoc >/dev/null 2>&1; then \
		$(CHMOD) +x doc/generate-man.sh; \
		cd doc && ./generate-man.sh; \
	else \
		echo "Warning: pandoc not found, skipping man page generation"; \
		echo "Install: pkg install hs-pandoc"; \
	fi

install: build
	$(MKDIR) $(DESTDIR)$(BINDIR)
	$(INSTALL) -m 755 pgmountd $(DESTDIR)$(BINDIR)/
	$(INSTALL) -m 755 pgmount $(DESTDIR)$(BINDIR)/
	$(INSTALL) -m 755 pgumount $(DESTDIR)$(BINDIR)/
	$(INSTALL) -m 755 pginfo $(DESTDIR)$(BINDIR)/
	$(MKDIR) $(DESTDIR)$(DATADIR)/examples/pgmount
	$(INSTALL) -m 644 config.example.yml $(DESTDIR)$(DATADIR)/examples/pgmount/
	@if [ -d doc/man ]; then \
		$(MKDIR) $(DESTDIR)$(MANDIR)/man8; \
		$(INSTALL) -m 644 doc/man/*.8 $(DESTDIR)$(MANDIR)/man8/; \
		echo "Man pages installed"; \
	fi

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/pgmountd
	rm -f $(DESTDIR)$(BINDIR)/pgmount
	rm -f $(DESTDIR)$(BINDIR)/pgumount
	rm -f $(DESTDIR)$(BINDIR)/pginfo
	rm -rf $(DESTDIR)$(DATADIR)/examples/pgmount
	rm -f $(DESTDIR)$(MANDIR)/man8/pgmountd.8
	rm -f $(DESTDIR)$(MANDIR)/man8/pgmount.8
	rm -f $(DESTDIR)$(MANDIR)/man8/pgumount.8
	rm -f $(DESTDIR)$(MANDIR)/man8/pginfo.8

clean:
	rm -f $(BINARIES)
	rm -rf doc/man
	$(GO) clean

test:
	$(GO) test -v ./...

format:
	$(GO) fmt ./...

lint:
	golangci-lint run

.PHONY: help
help:
	@echo "PGMount - Automounter for PGSD/FreeBSD/GhostBSD"
	@echo ""
	@echo "Targets:"
	@echo "  all      - Download dependencies and build all binaries (default)"
	@echo "  build    - Build all binaries"
	@echo "  deps     - Download Go dependencies"
	@echo "  man      - Generate man pages (requires pandoc)"
	@echo "  install  - Install binaries, documentation, and man pages"
	@echo "  uninstall- Remove installed files"
	@echo "  clean    - Remove built binaries and generated files"
	@echo "  test     - Run tests"
	@echo "  format   - Format source code"
	@echo "  lint     - Run linter"
	@echo ""
	@echo "Variables:"
	@echo "  PREFIX   - Installation prefix (default: /usr/local)"
	@echo "  DESTDIR  - Destination directory for staged installs"
	@echo ""
	@echo "First time building? Just run: make"
