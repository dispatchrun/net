// Package ttrpc provides an example of ttrpc client and server running in a
// program compiled to GOOS=wasip1
//
// The protoc compiler must be installed in order to build the tests in this
// package, as well as the following extension:
//
//	$ go install github.com/containerd/ttrpc/cmd/protoc-gen-go-ttrpc@latest
//
// For the full documentation on how to use ttrpc, see:
// https://github.com/containerd/ttrpc/blob/main/PROTOCOL.md
//
// When compiling to other targets than GOOS=wasip1, importing this package has
// no effect.
package ttrpc
