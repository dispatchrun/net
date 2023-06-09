// Package http exists only to modify the dial function of the default http
// transport, allowing all clients that rely on it to establish output
// connections when compiled to GOOS=wasip1.
//
// This package is intended to be imported as a nameless package with:
//
//	import (
//		_ "github.com/stealthrocket/net/http"
//	)
//
// Note that only the default transport (http.DefaultTransport) can be altered,
// other instances of http.Transport created by the program will default to use
// the standard library's net package, which does not have the ability to open
// network connections. Programs that create new transports must configure the
// dial function by setting the DialContext field to wasip1.DialContext.
//
// When compiling to other targets than GOOS=wasip1, importing this package has
// no effect.
package http
