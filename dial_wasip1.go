package net

import (
	"fmt"
	"net"
	"os"

	"github.com/stealthrocket/net/syscall"
)

// Conn is a generic stream-oriented network connection.
type Conn = net.Conn

// Dial connects to the address on the named network.
func Dial(network, address string) (Conn, error) {
	family, addr, err := lookupAddr("dial", network, address)
	if err != nil {
		return nil, err
	}
	return dialAddr(family, addr)
}

func dialAddr(family int, addr syscall.Sockaddr) (Conn, error) {
	fd, err := syscall.Socket(family, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, err
	}

	if err := syscall.SetNonblock(fd, true); err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("SetNonblock: %w", err)
	}

	var inProgress bool
	switch err := syscall.Connect(fd, addr); err {
	case nil:
	case syscall.EINPROGRESS:
		inProgress = true
	default:
		syscall.Close(fd)
		return nil, fmt.Errorf("Connect: %w", err)
	}

	f := os.NewFile(uintptr(fd), "")
	defer f.Close()

	if inProgress {
		rawConn, err := f.SyscallConn()
		if err != nil {
			return nil, err
		}
		rawConnErr := rawConn.Write(func(fd uintptr) bool {
			var value int
			value, err = syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_ERROR)
			if err != nil {
				return true // done
			}
			switch syscall.Errno(value) {
			case syscall.EINPROGRESS, syscall.EALREADY, syscall.EINTR:
				return false // continue
			case syscall.EISCONN:
				err = nil
				return true
			case syscall.Errno(0):
				// The net poller can wake up spuriously. Check that we are
				// are really connected.
				_, err := syscall.Getpeername(int(fd))
				return err == nil
			default:
				return true
			}
		})
		if err != nil {
			return nil, err
		} else if rawConnErr != nil {
			return nil, err
		}
	}

	c, err := net.FileConn(f)
	if err != nil {
		return nil, err
	}

	// TODO: get local+peer address; wrap FileConn to implement LocalAddr() and RemoteAddr()

	return c, nil
}
