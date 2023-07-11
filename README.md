[![Build](https://github.com/stealthrocket/net/actions/workflows/build.yml/badge.svg)](https://github.com/stealthrocket/net/actions/workflows/build.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/stealthrocket/net.svg)](https://pkg.go.dev/github.com/stealthrocket/net)
[![Apache 2 License](https://img.shields.io/badge/license-Apache%202-blue.svg)](LICENSE)

# net

This library provides `net.Dial` and `net.Listen` functions for
[`GOOS=wasip1`][wasip1].

Applications built with this library are compatible with [WasmEdge][wasmedge]
and [wasi-go][wasi-go] such as [Timecraft][timecraft].

[gotip]:     https://pkg.go.dev/golang.org/dl/gotip
[timecraft]: https://github.com/stealthrocket/timecraft
[wasi-go]:   https://github.com/stealthrocket/wasi-go
[wasip1]:    https://tip.golang.org/doc/go1.21#wasip1
[wasmedge]:  https://github.com/WasmEdge/WasmEdge

_Note: GOOS=wasip1 will be released in Go 1.21. At this time you would need to
use [gotip][gotip] to get the latest version of the Go compiler in order to use
this package._

## Motivation

The WASI preview 1 specification has partial support for socket networking,
preventing a large class of Go applications from running when compiled to
WebAssembly with `GOOS=wasip1`. Extensions to the base specifications have been
implemented by runtimes to enable a wider range of programs to be run as
WebAssembly modules.

This package aims to offset Go applications built with `GOOS=wasip1` the
opportunity to leverage those WASI extensions, by providing high level functions
similar to those found in the standard `net` package to create network clients
and servers.

## Configuration

Where possible, the package offers the ability to automatically configure the
network stack via `init` functions called on package imports. This model is
currently supported for `http` and `mysql` with those imports:

```go
import _ "github.com/stealthrocket/net/http"
```
```go
import _ "github.com/stealthrocket/net/mysql"
```

When imported, those packages alter the default configuration to install a
dialer function implemented on top of the WASI socket extensions. When compiled
to other targets, the import of those packages does nothing.

## Dialing

Packages implementing network clients for various protocols usually support
configuration through the installation of an alternative dial function allowing
the application to customize how network connections are established.

The `wasip1` sub-package provides dial functions matching the signature of those
implemented in the standard `net` package to integrate with those configuration
mechanisms.

The sub-modules contain examples of how to configure popular Go libraries to
leverage the dial functions of `wasip1`. Here is an example for a Redis client:

```go
client := redis.NewClient(&redis.Options{
	Addr:   "localhost:6379",
	Dialer: wasip1.DialContext, // change the dial function to use socket extensions
})
```

## Listening

Network servers can be created using the `wasip1.Listen` function, which mimics
the signature of `net.Listen` but uses WASI socket extensions to create the
`net.Listener`.

For example, a program compiled to `GOOS=wasip1` can create a http server by
first constructing a listener and passing it to the server's `Serve` method:

```go
import (
    "net/http"

    "github.com/stealthrocket/net/wasip1"
)

func main() {
    listener, err := wasip1.Listen("tcp", "127.0.0.1:3000")
    if err != nil {
        ...
    }
    server := &http.Server{
        ...
    }
    if err := server.Serve(listener); err != nil {
        ...
    }
}
```

Note that using convenience functions like `http.ListenAndServe` will not
work since they are hardcoded to depend on the standard `net` package.

## Name Resolution

There are two methods available for resolving a set of IP addresses for a
hostname.

### Pure Go Resolver

The pure Go name resolver is the default for `GOOS=wasip1`.

All you need is the following import somewhere in your application:

```go
import _ "github.com/stealthrocket/net/wasip1"
```

The library will then automatically configure the `net.DefaultResolver`.

You'll then be able to use the lookup functions from the standard
library (e.g. `net.LookupIP(host)`).

### getaddrinfo

The `sock_getaddrinfo` host function is used to implement name resolution.
To use this method, compile the library with the `getaddrinfo` build tag.

When using this method, the standard library resolver **will not work**; you
cannot use `net.DefaultResolver`, `net.LookupIP`, etc.

Note that `sock_getaddrinfo` may block.
