//go:build wasip1

package websocket_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/stealthrocket/net/http"
	"github.com/stealthrocket/net/wasip1"
	ws "nhooyr.io/websocket"
)

func TestWebSocket(t *testing.T) {
	// We must use NewUnstartedServer instead of NewServer otherwise the server
	// attempts to open a listener but it does not know that it has to use the
	// wasip1.Listen function.
	server := httptest.NewUnstartedServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := ws.Accept(w, r, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer conn.Close(ws.StatusNormalClosure, "OK")
			conn.Write(r.Context(), ws.MessageText, []byte("Hello, World!\n"))
		}),
	)
	defer server.Close()

	l, err := wasip1.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	server.Listener = l
	server.Start()

	ctx := context.Background()

	if deadline, ok := t.Deadline(); ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
	}

	addr := server.URL
	addr = strings.TrimPrefix(addr, "http://")
	addr = "ws://" + addr

	// Open a WebSocket connection using the default configuration, which
	// creates a default http.Client relying on http.DefaultTransport that the
	// import of github.com/stealthrockte/net/http/ package configured to use
	// wasip1.DialContext.
	conn, _, err := ws.Dial(ctx, addr, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ws.StatusNormalClosure, "OK")

	_, msg, err := conn.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(msg) != "Hello, World!\n" {
		t.Errorf("wrong websocket message received: %q", msg)
	}
}
