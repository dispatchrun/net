// Package grpc provides an example of gRPC client and server running in a
// program compiled to GOOS=wasip1
//
// The protoc compiler must be installed in order to build the tests in this
// package, as well as the following extension:
//
//	$ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
//	$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
//
// For the full documentation on how to use gRPC, see:
// https://grpc.io/docs/languages/go/quickstart/
//
// When compiling to other targets than GOOS=wasip1, importing this package has
// no effect.
package grpc
