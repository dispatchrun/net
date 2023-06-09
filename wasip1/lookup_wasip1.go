//go:build wasip1 && !purego

package wasip1

import (
	"net"
	"os"

	"github.com/stealthrocket/net/internal/syscall"
)

func lookupAddr(op, network, address string) (net.Addr, error) {
	var hints syscall.AddrInfo
	switch network {
	case "tcp", "tcp4", "tcp6":
		hints.SocketType = syscall.SOCK_STREAM
		hints.Protocol = syscall.IPPROTO_TCP
	case "udp", "udp4", "udp6":
		hints.SocketType = syscall.SOCK_DGRAM
		hints.Protocol = syscall.IPPROTO_UDP
	case "unix", "unixgram":
		return &net.UnixAddr{Name: address, Net: network}, nil
	default:
		return nil, net.UnknownNetworkError(network)
	}
	switch network {
	case "tcp", "udp":
		hints.Family = syscall.AF_UNSPEC
	case "tcp4", "udp4":
		hints.Family = syscall.AF_INET
	case "tcp6", "udp6":
		hints.Family = syscall.AF_INET6
	}
	hostname, service, err := net.SplitHostPort(address)
	if err != nil {
		return nil, net.InvalidAddrError(address)
	}
	if op == "listen" && hostname == "" {
		hints.Flags |= syscall.AI_PASSIVE
	}

	results := make([]syscall.AddrInfo, 16)
	n, err := syscall.Getaddrinfo(hostname, service, hints, results)
	if err != nil {
		addr := &netAddr{network, address}
		return nil, newOpError(op, addr, os.NewSyscallError("getaddrinfo", err))
	}
	results = results[:n]
	for _, r := range results {
		var ip net.IP
		var port int
		switch a := r.Address.(type) {
		case *syscall.SockaddrInet4:
			ip = a.Addr[:]
			port = a.Port
		case *syscall.SockaddrInet6:
			ip = a.Addr[:]
			port = a.Port
		}
		switch network {
		case "tcp", "tcp4", "tcp6":
			return &net.TCPAddr{IP: ip, Port: port}, nil
		case "udp", "udp4", "udp6":
			return &net.UDPAddr{IP: ip, Port: port}, nil
		}
	}
	return nil, &net.DNSError{
		Err:        "lookup failed",
		Name:       hostname,
		IsNotFound: true,
	}
}
