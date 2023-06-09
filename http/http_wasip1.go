//go:build wasip1

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
