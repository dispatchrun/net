// Package http exists only to modify the dial function of the default http
// transport, allowing all clients that rely on it to establish output
// connections when compiled to GOOS=wasip1.
package http

import (
	"net/http"

	"github.com/stealthrocket/net/wasip1"
)

func init() {
	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		t.DialContext = wasip1.DialContext
	}
}
