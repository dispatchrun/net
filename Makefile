.PHONY: test lint wasirun

GOPATH ?= $(shell $(GO) env GOPATH)
wasirun = $(GOPATH)/bin/wasirun

packages.dir = $(wildcard */)
packages.test = $(packages.dir:/=.test)

test: wasirun $(packages.test)
	@for pkg in $(packages.test); do \
		tmp=$$(mktemp); \
		$(wasirun) $$pkg > $$tmp; \
		if (($$?)); then cat $$tmp; exit 1; else printf "ok\tgithub.com/stealthrocket/net/$$pkg\n"; fi \
	done

wasirun: $(wasirun)

$(wasirun):
	$(GO) install github.com/stealthrocket/wasi-go/cmd/wasirun@latest

%.test: %/
	cd $< && GOARCH=wasm GOOS=wasip1 $(GO) test -c -o ../$(notdir $@)
