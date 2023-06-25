//go:build wasip1

package wasip1

// Unix sockets are not yet supported for GOOS=wasip1 because there is no
// mechanism to create a *net.UnixConn or *net.UnixListener from an *os.File
// due to WASI preview 1 not having the concept of unix sockets and only
// having file types for datagram and stream sockets, which are mapped to
// UDP and TCP sockets by the net package.
//
// We emulate unix sockets to provide minimum support when calling Dial with
// a "unix" network. The downside is that connections returned by the wasip1
// dial functions are not actually of type *net.UnixConn, applications that
// used type assertions to dynamically discover the connection type will not
// be compatible with this approach. The same limitation applies to listeners.

import (
	"net"
	"syscall"
)

type unixConn struct {
	net.Conn
	laddr net.UnixAddr
	raddr net.UnixAddr
}

func (c *unixConn) LocalAddr() net.Addr {
	return &c.laddr
}

func (c *unixConn) RemoteAddr() net.Addr {
	return &c.raddr
}

func (c *unixConn) CloseRead() error {
	if cr, ok := c.Conn.(closeReader); ok {
		return cr.CloseRead()
	}
	return &net.OpError{
		Op:     "close",
		Net:    "unix",
		Source: c.LocalAddr(),
		Err:    syscall.ENOTSUP,
	}
}

func (c *unixConn) CloseWrite() error {
	if cw, ok := c.Conn.(closeWriter); ok {
		return cw.CloseWrite()
	}
	return &net.OpError{
		Op:     "close",
		Net:    "unix",
		Source: c.LocalAddr(),
		Err:    syscall.ENOTSUP,
	}
}

type closeReader interface {
	CloseRead() error
}

type closeWriter interface {
	CloseWrite() error
}

var (
	_ closeReader = (*net.UnixConn)(nil)
	_ closeWriter = (*net.UnixConn)(nil)

	_ closeReader = (*unixConn)(nil)
	_ closeWriter = (*unixConn)(nil)
)
