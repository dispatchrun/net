//go:build wasip1

package wasip1

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
)

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

func family(addr net.Addr) int {
	var ip net.IP
	switch a := addr.(type) {
	case *net.IPAddr:
		ip = a.IP
	case *net.TCPAddr:
		ip = a.IP
	case *net.UDPAddr:
		ip = a.IP
	case *net.UnixAddr:
		return AF_UNIX
	}
	if ip.To4() != nil {
		return AF_INET
	} else if len(ip) == net.IPv6len {
		return AF_INET6
	}
	return AF_INET
}

func socketType(addr net.Addr) (int, error) {
	switch addr.Network() {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
		return SOCK_STREAM, nil
	case "udp", "udp4", "udp6", "unixgram":
		return SOCK_DGRAM, nil
	default:
		return -1, syscall.EPROTOTYPE
	}
}

func socketAddress(addr net.Addr) (sockaddr, error) {
	var ip net.IP
	var port int
	switch a := addr.(type) {
	case *net.IPAddr:
		ip = a.IP
	case *net.TCPAddr:
		ip, port = a.IP, a.Port
	case *net.UDPAddr:
		ip, port = a.IP, a.Port
	case *net.UnixAddr:
		return &sockaddrUnix{name: a.Name}, nil
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		return &sockaddrInet4{addr: ([4]byte)(ipv4), port: uint32(port)}, nil
	} else if len(ip) == net.IPv6len {
		return &sockaddrInet6{addr: ([16]byte)(ip), port: uint32(port)}, nil
	} else {
		return nil, &net.AddrError{
			Err:  "unsupported address type",
			Addr: addr.String(),
		}
	}
}

// In Go 1.21, the net package cannot initialize the local and remote addresses
// of network connections. For this reason, we use this function to retreive the
// addresses and return a wrapped net.Conn with LocalAddr/RemoteAddr implemented.
func makeConn(c net.Conn) (net.Conn, error) {
	syscallConn, ok := c.(syscall.Conn)
	if !ok {
		return c, nil
	}
	rawConn, err := syscallConn.SyscallConn()
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("syscall.Conn.SyscallConn: %w", err)
	}
	rawConnErr := rawConn.Control(func(fd uintptr) {
		var addr sockaddr
		var peer sockaddr

		if addr, err = getsockname(int(fd)); err != nil {
			err = os.NewSyscallError("getsockname", err)
			return
		}

		if peer, err = getpeername(int(fd)); err != nil {
			err = os.NewSyscallError("getpeername", err)
			return
		}

		if _, unix := addr.(*sockaddrUnix); unix {
			c = &unixConn{Conn: c}
		}

		setNetAddr(SOCK_STREAM, c.LocalAddr(), addr)
		setNetAddr(SOCK_STREAM, c.RemoteAddr(), peer)
	})
	if err == nil {
		err = rawConnErr
	}
	if err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

func setNetAddr(sotype int, dst net.Addr, src sockaddr) {
	switch a := dst.(type) {
	case *net.IPAddr:
		a.IP, _ = sockaddrIPAndPort(src)
	case *net.TCPAddr:
		a.IP, a.Port = sockaddrIPAndPort(src)
	case *net.UDPAddr:
		a.IP, a.Port = sockaddrIPAndPort(src)
	case *net.UnixAddr:
		switch sotype {
		case SOCK_STREAM:
			a.Net = "unix"
		case SOCK_DGRAM:
			a.Net = "unixgram"
		}
		a.Name = sockaddrName(src)
	}
}

func sockaddrName(addr sockaddr) string {
	switch a := addr.(type) {
	case *sockaddrUnix:
		return a.name
	default:
		return ""
	}
}

func sockaddrIPAndPort(addr sockaddr) (net.IP, int) {
	switch a := addr.(type) {
	case *sockaddrInet4:
		return net.IP(a.addr[:]), int(a.port)
	case *sockaddrInet6:
		return net.IP(a.addr[:]), int(a.port)
	default:
		return nil, 0
	}
}

func setNonBlock(fd int) error {
	if err := syscall.SetNonblock(fd, true); err != nil {
		return os.NewSyscallError("setnonblock", err)
	}
	return nil
}

func setReuseAddress(fd int) error {
	if err := setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, 1); err != nil {
		// The runtime may not support the option; if that's the case and the
		// address is already in use, binding the socket will fail and we will
		// report the error then.
		switch {
		case errors.Is(err, syscall.ENOPROTOOPT):
		case errors.Is(err, syscall.EINVAL):
		default:
			return os.NewSyscallError("setsockopt", err)
		}
	}
	return nil
}
