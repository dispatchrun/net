.PHONY: test lint wasirun

GOPATH ?= $(shell $(GO) env GOPATH)
wasirun = $(GOPATH)/bin/wasirun

wasip1.test: go.mod $(wildcard wasip1/*.go)
	GOARCH=wasm GOOS=wasip1 $(GO) test -c ./wasip1

test: wasirun wasip1.test
	$(wasirun) wasip1.test -test.v

wasirun: $(wasirun)

$(wasirun):
	$(GO) install github.com/stealthrocket/wasi-go/cmd/wasirun@latest
