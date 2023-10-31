PREFIX := /usr/local

build: test 127
.PHONY: build

127:
	go build -o 127

check: test lint
.PHONY: check

test:
	go test -race -cover ./...
.PHONY: test

lint:
	golangci-lint run
.PHONY: lint

install:
	install -D -m 0755 -t $(DESTDIR)$(PREFIX)/bin 127
	install -D -m 0644 -t $(DESTDIR)$(PREFIX)/share/doc/127 README.md LICENSE
.PHONY: install

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/127
	rm -rf $(DESTDIR)$(PREFIX)/share/doc/127
.PHONY: uninstall

snapshot:
	goreleaser release --clean --snapshot
.PHONY: snapshot

dist: check
	goreleaser release --clean --skip=announce,publish,validate
.PHONY: dist

release: check
	goreleaser release --clean
.PHONY: release

clean:
	go clean ./...
	rm -rf ./dist
.PHONY: clean
