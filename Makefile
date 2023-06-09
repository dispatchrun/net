.PHONY: test lint wasirun

GOPATH ?= $(shell $(GO) env GOPATH)

wasip1.test: go.mod $(wildcard wasip1/*.go)
	GOARCH=wasm GOOS=wasip1 $(GO) test -c ./wasip1

test: wasirun wasip1.test
	wasirun wasip1.test -test.v

wasirun: $(GOPATH)/bin/wasirun

$(GOPATH)/bin/wasirun:
	$(GO) install github.com/stealthrocket/wasi-go/cmd/wasirun@latest
