PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
MANDIR ?= $(PREFIX)/share/man

.PHONY: build install install-man man man-ru man-en test uninstall snapshot release-cross

build:
	go build -o kiri .

install: build install-man
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 kiri $(DESTDIR)$(BINDIR)/

install-man:
	install -d $(DESTDIR)$(MANDIR)/man1 $(DESTDIR)$(MANDIR)/ru/man1
	install -m 644 doc/en/kiri.1 $(DESTDIR)$(MANDIR)/man1/
	install -m 644 doc/ru/kiri.1 $(DESTDIR)$(MANDIR)/ru/man1/

man: man-en

man-en:
	man -l doc/en/kiri.1

man-ru:
	man -l doc/ru/kiri.1

test:
	go test ./...

# Local snapshot: linux/amd64 only (native gcc + CGO).
snapshot:
	goreleaser release --snapshot --clean

# All targets (linux + macOS, amd64 + arm64) via goreleaser-cross.
release-cross:
	docker run --rm --privileged \
		--user $(shell id -u):$(shell id -g) \
		-v "$(CURDIR)":/workspace -w /workspace \
		-e GORELEASER_CROSS=1 \
		ghcr.io/goreleaser/goreleaser-cross:v1.26.3-v2.16.0 \
		release --snapshot --clean

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/kiri
	rm -f $(DESTDIR)$(MANDIR)/man1/kiri.1
	rm -f $(DESTDIR)$(MANDIR)/ru/man1/kiri.1