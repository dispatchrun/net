//go:build wasip1

package http_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/stealthrocket/net/http"
	"github.com/stealthrocket/net/wasip1"
)

func TestHTTP(t *testing.T) {
	// We must use NewUnstartedServer instead of NewServer otherwise the server
	// attempts to open a listener but it does not know that it has to use the
	// wasip1.Listen function.
	server := httptest.NewUnstartedServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "Hello, World!\n")
		}),
	)
	defer server.Close()

	l, err := wasip1.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	server.Listener = l
	server.Start()

	r, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "Hello, World!\n" {
		t.Errorf("wrong http response received: %q", b)
	}
}
