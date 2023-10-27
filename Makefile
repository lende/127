PREFIX := /usr/local

.PHONY: build
all: test 127

127:
	go build -o 127

.PHONY: build
build: 127

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: install
install:
	install -D -m 0755 -t $(DESTDIR)$(PREFIX)/bin 127
	install -D -m 0644 -t $(DESTDIR)$(PREFIX)/share/doc/127 README.md LICENSE

.PHONY: uninstall
uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/127
	rm -rf $(DESTDIR)$(PREFIX)/share/doc/127

.PHONY: clean
clean:
	go clean ./...
