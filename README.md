# net

This library provides `net.Dial` and `net.Listen` functions
for `GOOS=wasip1`. It uses the WasmEdge sockets extension to WASI
preview 1.

Applications built with this library are compatible with WasmEdge
and [stealthrocket/wasi-go](https://github.com/stealthrocket/wasi-go).

## Dialing

The library will automatically configure the default HTTP transport
to use the `Dial` function from this library. To make outbound HTTP 
connections you just need the following import somewhere:

```go
import _ "github.com/stealthrocket/net"
```

To connect to databases, there's usually a way to pass in a custom `Dial`
function.

For example, to connect to MySQL:

```go
import (
    "context"

    "github.com/go-sql-driver/mysql"
    "github.com/stealthrocket/net"
)

func init() {
    for _, network := range []string{"tcp", "tcp4", "tcp6"} {
        mysql.RegisterDialContext(network, func(ctx context.Context, addr string) (net.Conn, error) {
            return net.Dial(network, addr)
        })
    }
}

func main() {
    db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/database")
}
```

For example, to connect to Redis:

```go
import (
    "github.com/redis/go-redis/v9"
    "github.com/stealthrocket/net"
)

func main() {
    db := redis.NewClient(&redis.Options{
        Addr:   "127.0.0.1:6379",
        Dialer: net.DialContext,
    })
}
```

## Listening

HTTP servers can be created like so:

```go
import (
    "net/http"

    "github.com/stealthrocket/net"
)

func main() {
    listener, err := net.Listen("tcp", "127.0.0.1:8080")
    if err != nil {
        // TODO: handle listen error
    }
    server := &http.Server{
        // TODO: setup HTTP server
    }
    err = server.Serve(listener)
}
```

## Name Resolution

There are two methods available for resolving a set of IP addresses
for a hostname.

### getaddrinfo

The `sock_getaddrinfo` host function is used to implement name resolution.
This requires WasmEdge, or a WasmEdge compatible WASI layer
(e.g. [wasi-go](http://github.com/stealthrocket/wasi-go)).

When using this method, the standard library resolver **will not work**. You
_cannot_ use `net.DefaultResolver`, `net.LookupIP`, etc. with this approach
because the standard library does not allow us to patch it with an alternative
implementation.

Note that `sock_getaddrinfo` may block!

### Pure Go Resolver

The pure Go name resolver is not currently enabled for GOOS=wasip1.

The following series of CLs will change this: https://go-review.googlesource.com/c/go/+/500576.
This will hopefully land in Go v1.22 in ~February 2024.

If you're using a version of Go that has the CL's included, you can
instruct this library to use the pure Go resolver by including the
`purego` build tag.

The library will then automatically configure the `net.DefaultResolver`.
All you need is the following import somewhere in your application:

```go
import _ "github.com/stealthrocket/net"
```

You should then be able to use the lookup functions from the standard
library (e.g. `net.LookupIP(host)`).
