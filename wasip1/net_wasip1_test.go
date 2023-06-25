//go:build wasip1

package wasip1_test

import (
	"net"
	"path/filepath"
	"testing"

	"github.com/stealthrocket/net/wasip1"
	"golang.org/x/net/nettest"
)

func TestConn(t *testing.T) {
	tests := []struct {
		network string
		address string
	}{
		{
			network: "tcp",
			address: ":0",
		},
		{
			network: "tcp4",
			address: ":0",
		},
		{
			network: "tcp6",
			address: ":0",
		},
		{
			network: "unix",
			address: ":0",
		},
	}

	for _, test := range tests {
		t.Run(test.network, func(t *testing.T) {
			nettest.TestConn(t, func() (c1, c2 net.Conn, stop func(), err error) {
				network := test.network
				address := test.address

				switch network {
				case "unix":
					address = filepath.Join(t.TempDir(), address)
				}

				l, err := wasip1.Listen(network, address)
				if err != nil {
					return nil, nil, nil, err
				}
				defer l.Close()

				conns := make(chan net.Conn, 1)
				errch := make(chan error, 1)
				go func() {
					c, err := l.Accept()
					if err != nil {
						errch <- err
					} else {
						conns <- c
					}
				}()

				dialer := &wasip1.Dialer{}
				dialer.Deadline, _ = t.Deadline()

				laddr := l.Addr()
				c1, err = dialer.Dial(laddr.Network(), laddr.String())
				if err != nil {
					return nil, nil, nil, err
				}

				select {
				case c2 := <-conns:
					return c1, c2, func() { c1.Close(); c2.Close() }, nil
				case err := <-errch:
					c1.Close()
					return nil, nil, nil, err
				}
			})
		})
	}
}
