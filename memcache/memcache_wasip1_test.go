//go:build wasip1

package memcache_test

import (
	"testing"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/stealthrocket/net/wasip1"
)

func TestMemcache(t *testing.T) {
	client := memcache.New("localhost:11211")
	defer client.Close()
	// Change the dial function so the client uses the WASI socket extensions
	// missing from Go 1.21.
	client.DialContext = wasip1.DialContext

	if err := client.Ping(); err != nil {
		t.Fatal(err)
	}
}
