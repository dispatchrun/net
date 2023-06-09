package wasip1_test

import (
	"net"
	"testing"

	"github.com/stealthrocket/net/wasip1"
	"golang.org/x/net/nettest"
)

func TestConn(t *testing.T) {
	// TODO: for now only the TCP tests pass due to limitations in Go 1.21, see:
	// https://github.com/golang/go/blob/39effbc105f5c54117a6011af3c48e3c8f14eca9/src/net/file_wasip1.go#L33-L55
	//
	// Once https://go-review.googlesource.com/c/go/+/500578 is merged, we will
	// be able to test udp and unix networks as well.
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
	}

	for _, test := range tests {
		t.Run(test.network, func(t *testing.T) {
			nettest.TestConn(t, func() (c1, c2 net.Conn, stop func(), err error) {
				l, err := wasip1.Listen(test.network, test.address)
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

				address := l.Addr()
				c1, err = dialer.Dial(address.Network(), address.String())
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
