package net

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/stealthrocket/net/syscall"
)

func init() {
	net.DefaultResolver.Dial = DialContext
}

// Conn is a generic stream-oriented network connection.
type Conn = net.Conn

// Dial connects to the address on the named network.
func Dial(network, address string) (Conn, error) {
	addr, err := lookupAddr("dial", network, address)
	if err != nil {
		return nil, err
	}
	return dialAddr(addr)
}

// DialContext is a variant of Dial that accepts a context.
func DialContext(ctx context.Context, network, address string) (Conn, error) {
	_ = ctx // TODO
	return Dial(network, address)
}

func dialAddr(addr net.Addr) (Conn, error) {
	proto := family(addr)
	sotype := socketType(addr)

	fd, err := syscall.Socket(proto, sotype, 0)
	if err != nil {
		return nil, err
	}

	if err := syscall.SetNonblock(fd, true); err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("SetNonblock: %w", err)
	}

	if sotype == syscall.SOCK_DGRAM && proto != syscall.AF_UNIX {
		if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1); err != nil {
			syscall.Close(fd)
			return nil, err
		}
	}

	var inProgress bool
	switch err := syscall.Connect(fd, socketAddress(addr)); err {
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
			case syscall.EINPROGRESS, syscall.EINTR:
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

func family(addr net.Addr) int {
	var ip net.IP
	switch a := addr.(type) {
	case *net.UnixAddr:
		return syscall.AF_UNIX
	case *net.TCPAddr:
		ip = a.IP
	case *net.UDPAddr:
		ip = a.IP
	case *net.IPAddr:
		ip = a.IP
	}
	switch len(ip) {
	case net.IPv4len:
		return syscall.AF_INET
	case net.IPv6len:
		return syscall.AF_INET6
	default:
		panic("invalid IP address")
	}
}

func socketType(addr net.Addr) int {
	switch addr.Network() {
	case "tcp", "unix":
		return syscall.SOCK_STREAM
	case "udp", "unixgram":
		return syscall.SOCK_DGRAM
	default:
		panic("not implemented")
	}
}

func socketAddress(addr net.Addr) syscall.Sockaddr {
	var ip net.IP
	var port int
	switch a := addr.(type) {
	case *net.UnixAddr:
		return &syscall.SockaddrUnix{Name: a.Name}
	case *net.TCPAddr:
		ip, port = a.IP, a.Port
	case *net.UDPAddr:
		ip, port = a.IP, a.Port
	case *net.IPAddr:
		ip = a.IP
	}
	switch len(ip) {
	case net.IPv4len:
		return &syscall.SockaddrInet4{Addr: ([4]byte)(ip), Port: port}
	case net.IPv6len:
		return &syscall.SockaddrInet6{Addr: ([16]byte)(ip), Port: port}
	default:
		panic("invalid IP address")
	}
}
