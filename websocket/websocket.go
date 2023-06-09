// Package websocket provides an example of a WebSocket client and server
// running in a program compiled to GOOS=wasip1.
//
// The test uses https://pkg.go.dev/nhooyr.io/websocket and relies on the
// github.com/stealthrocket/net/http package to configure the default http
// transport, as well as github.com/stealthrocket/wasip1 to create the server
// listener accepting connections.
//
// When compiling to other targets than GOOS=wasip1, importing this package has
// no effect.
package websocket
