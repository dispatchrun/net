.PHONY: test lint wasirun

GOPATH ?= $(shell $(GO) env GOPATH)
wasirun = $(GOPATH)/bin/wasirun

packages.dir = $(wildcard */)
packages.test = $(packages.dir:/=.test)

test: wasirun $(packages.test)
	for pkg in $(packages.test); do $(wasirun) $$pkg -test.v || exit 1; done

wasirun: $(wasirun)

$(wasirun):
	$(GO) install github.com/stealthrocket/wasi-go/cmd/wasirun@latest

%.test: %/
	cd $< && GOARCH=wasm GOOS=wasip1 $(GO) test -c -o ../$(notdir $@)
