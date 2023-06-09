.PHONY: test

go ?= go

test: wasip1.test
	wasirun wasip1.test -test.v

wasip1.test:
	GOARCH=wasm GOOS=wasip1 $(go) test -c ./wasip1
