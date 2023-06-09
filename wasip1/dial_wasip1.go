package wasip1

import (
	"context"
	"net"
	"os"
	"syscall"
)

// Dial connects to the address on the named network.
func Dial(network, address string) (net.Conn, error) {
	addr, err := lookupAddr("dial", network, address)
	if err != nil {
		addr := &netAddr{network, address}
		return nil, dialErr(addr, err)
	}
	conn, err := dialAddr(addr)
	if err != nil {
		return nil, dialErr(addr, err)
	}
	return conn, nil
}

// DialContext is a variant of Dial that accepts a context.
func DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	select {
	case <-ctx.Done():
		addr := &netAddr{network, address}
		return nil, dialErr(addr, context.Cause(ctx))
	default:
		return Dial(network, address)
	}
}

func dialErr(addr net.Addr, err error) error {
	return newOpError("dial", addr, err)
}

func dialAddr(addr net.Addr) (net.Conn, error) {
	proto := family(addr)
	sotype, err := socketType(addr)
	if err != nil {
		return nil, os.NewSyscallError("socket", err)
	}
	fd, err := socket(proto, sotype, 0)
	if err != nil {
		return nil, os.NewSyscallError("socket", err)
	}

	if err := syscall.SetNonblock(fd, true); err != nil {
		syscall.Close(fd)
		return nil, os.NewSyscallError("setnonblock", err)
	}
	if sotype == SOCK_DGRAM && proto != AF_UNIX {
		if err := setsockopt(fd, SOL_SOCKET, SO_BROADCAST, 1); err != nil {
			syscall.Close(fd)
			return nil, os.NewSyscallError("setsockopt", err)
		}
	}

	connectAddr, err := socketAddress(addr)
	if err != nil {
		return nil, os.NewSyscallError("connect", err)
	}
	var inProgress bool
	switch err := connect(fd, connectAddr); err {
	case nil:
	case syscall.EINPROGRESS:
		inProgress = true
	default:
		syscall.Close(fd)
		return nil, os.NewSyscallError("connect", err)
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
			value, err = getsockopt(int(fd), SOL_SOCKET, SO_ERROR)
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
				_, err := getpeername(int(fd))
				return err == nil
			default:
				return true
			}
		})
		if err == nil {
			err = rawConnErr
		}
		if err != nil {
			return nil, os.NewSyscallError("connect", err)
		}
	}

	c, err := net.FileConn(f)
	if err != nil {
		return nil, err
	}
	return makeConn(c)
}
