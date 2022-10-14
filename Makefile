DESTDIR =
PREFIX  =/usr/local
all:
clean:
install:

## -- license --
install: install-license
install-license: LICENSE
	@echo 'I share/doc/go-ustripe/LICENSE'
	@mkdir -p $(DESTDIR)$(PREFIX)/share/doc/go-ustripe
	@cp LICENSE $(DESTDIR)$(PREFIX)/share/doc/go-ustripe
## -- license --
## -- AUTO-GO --
all:     all-go
install: install-go
clean:   clean-go
all-go:
	@echo "B bin/ustripe     ./cmd/ustripe"
	@go build -o bin/ustripe     ./cmd/ustripe
install-go: all-go
	@install -d $(DESTDIR)$(PREFIX)/bin
	@echo I bin/ustripe
	@cp bin/ustripe     $(DESTDIR)$(PREFIX)/bin
clean-go:
	rm -f bin/ustripe
## -- AUTO-GO --
