package wasip1

import (
	"net"
	"os"

	"github.com/stealthrocket/net/internal/syscall"
)

// Listen announces on the local network address.
func Listen(network, address string) (net.Listener, error) {
	addr, err := lookupAddr("listen", network, address)
	if err != nil {
		addr := &netAddr{network, address}
		return nil, listenErr(addr, err)
	}
	lstn, err := listenAddr(addr)
	if err != nil {
		return nil, listenErr(addr, err)
	}
	return lstn, nil
}

func listenErr(addr net.Addr, err error) error {
	return newOpError("listen", addr, err)
}

func listenAddr(addr net.Addr) (net.Listener, error) {
	fd, err := syscall.Socket(family(addr), socketType(addr), 0)
	if err != nil {
		return nil, os.NewSyscallError("socket", err)
	}

	if err := syscall.SetNonblock(fd, true); err != nil {
		syscall.Close(fd)
		return nil, os.NewSyscallError("setnonblock", err)
	}
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		syscall.Close(fd)
		return nil, os.NewSyscallError("setsockopt", err)
	}

	listenAddr, err := socketAddress(addr)
	if err != nil {
		return nil, os.NewSyscallError("bind", err)
	}

	if err := syscall.Bind(fd, listenAddr); err != nil {
		syscall.Close(fd)
		return nil, os.NewSyscallError("bind", err)
	}

	const backlog = 64 // TODO: configurable?
	if err := syscall.Listen(fd, backlog); err != nil {
		syscall.Close(fd)
		return nil, os.NewSyscallError("listen", err)
	}

	f := os.NewFile(uintptr(fd), "")
	defer f.Close()

	l, err := net.FileListener(f)
	if err != nil {
		return nil, err
	}
	return &listener{l, addr}, nil
}

type listener struct {
	net.Listener
	addr net.Addr
}

func (l *listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	// TODO: get local+peer address; wrap Conn to implement LocalAddr() and RemoteAddr()
	return c, nil
}

func (l *listener) Addr() net.Addr {
	return l.addr
}
