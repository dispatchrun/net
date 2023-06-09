package wasip1

import (
	"context"
	"errors"
	"net"
	"net/http"
)

func dialResolverNotSupported(ctx context.Context, network, address string) (net.Conn, error) {
	// The net.Resolver type makes a call to net.DialUDP to determine which
	// resolved addresses are reachable, which does not go through its Dial
	// hook. As a result, it is unusable on GOOS=wasip1 because it fails
	// even when the Dial function is set because WASI preview 1 does not
	// have a mechanism for opening UDP sockets.
	//
	// Instead of having (often indirect) use of the net.Resolver crash, we
	// override the Dial function to error earlier in the resolver lifecycle
	// with an error which is more explicit to the end user.
	return nil, errors.New("net.Resolver not supported on GOOS=wasip1")
}

func init() {
	net.DefaultResolver.Dial = dialResolverNotSupported

	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		t.DialContext = DialContext
	}
}

func newOpError(op string, addr net.Addr, err error) error {
	return &net.OpError{
		Op:   op,
		Net:  addr.Network(),
		Addr: addr,
		Err:  err,
	}
}

type netAddr struct{ network, address string }

func (na *netAddr) Network() string { return na.address }
func (na *netAddr) String() string  { return na.address }
