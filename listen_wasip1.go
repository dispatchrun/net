package net

import (
	"fmt"
	"net"
	"os"

	"github.com/stealthrocket/net/syscall"
)

// A Listener is a generic network listener for stream-oriented protocols.
type Listener = net.Listener

// Listen announces on the local network address.
func Listen(network, address string) (Listener, error) {
	addr, err := lookupAddr("listen", network, address)
	if err != nil {
		return nil, err
	}
	return listenAddr(addr)
}

func listenAddr(addr net.Addr) (Listener, error) {
	fd, err := syscall.Socket(family(addr), socketType(addr), 0)
	if err != nil {
		return nil, fmt.Errorf("Socket: %w", err)
	}

	if err := syscall.SetNonblock(fd, true); err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("SetNonBlock: %w", err)
	}
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		syscall.Close(fd)
		return nil, err
	}

	if err := syscall.Bind(fd, socketAddress(addr)); err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("Bind: %w", err)
	}

	const backlog = 64 // TODO: configurable?
	if err := syscall.Listen(fd, backlog); err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("Listen: %w", err)
	}

	f := os.NewFile(uintptr(fd), "")
	defer f.Close()

	l, err := net.FileListener(f)
	if err != nil {
		return nil, err
	}

	// TODO: get local address; wrap FileListener to implement Addr().
	//  Wrap the net.Conn returned by Accept() to implement LocalAddr() and
	//  RemoteAddr()

	return l, err
}
