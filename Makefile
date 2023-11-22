PREFIX := /usr/local

LINT    = go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
RELEASE = go run github.com/goreleaser/goreleaser@v1.22.1 release --clean

VERSION != git describe --long --dirty

build: check 127
.PHONY: build

127:
	go build -o 127 -ldflags "-X main.version=$(VERSION)"

run:
	@go run .
.PHONY: run

version:
	@go run -ldflags "-X main.version=$(VERSION)" . -v
.PHONY: version

check: test lint
.PHONY: check

test:
	go test -race -cover ./...
.PHONY: test

lint:
	$(LINT) run
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
	$(RELEASE) --snapshot
.PHONY: snapshot

dist: check
	$(RELEASE) --skip=announce,publish,validate
.PHONY: dist

release: check
	$(RELEASE) --clean
.PHONY: release

clean:
	go clean ./...
	rm -rf ./dist
.PHONY: clean
