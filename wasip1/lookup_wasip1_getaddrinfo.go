//go:build wasip1 && getaddrinfo

package wasip1

import (
	"context"
	"net"
	"os"
	"strconv"
)

func lookupAddr(ctx context.Context, op, network, address string) ([]net.Addr, error) {
	var hints addrInfo

	switch network {
	case "tcp", "tcp4", "tcp6":
		hints.socketType = SOCK_STREAM
		hints.protocol = IPPROTO_TCP
	case "udp", "udp4", "udp6":
		hints.socketType = SOCK_DGRAM
		hints.protocol = IPPROTO_UDP
	case "unix", "unixgram":
		return []net.Addr{&net.UnixAddr{Name: address, Net: network}}, nil
	default:
		return nil, net.UnknownNetworkError(network)
	}

	switch network {
	case "tcp", "udp":
		hints.family = AF_UNSPEC
	case "tcp4", "udp4":
		hints.family = AF_INET
	case "tcp6", "udp6":
		hints.family = AF_INET6
	}

	hostname, service, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	if ip := net.ParseIP(hostname); ip != nil {
		hints.flags |= AI_NUMERICHOST
	}
	if _, err = strconv.Atoi(service); err == nil {
		hints.flags |= AI_NUMERICSERV
	}
	if op == "listen" && hostname == "" {
		hints.flags |= AI_PASSIVE
	}

	results := make([]addrInfo, 8)
	n, err := getaddrinfo(hostname, service, &hints, results)
	if err != nil {
		addr := &netAddr{network, address}
		return nil, newOpError(op, addr, os.NewSyscallError("getaddrinfo", err))
	}

	addrs := make([]net.Addr, 0, n)
	for _, r := range results[:n] {
		var ip net.IP
		var port int
		switch a := r.address.(type) {
		case *sockaddrInet4:
			ip = a.addr[:]
			port = int(a.port)
		case *sockaddrInet6:
			ip = a.addr[:]
			port = int(a.port)
		}
		switch network {
		case "tcp", "tcp4", "tcp6":
			addrs = append(addrs, &net.TCPAddr{IP: ip, Port: port})
		case "udp", "udp4", "udp6":
			addrs = append(addrs, &net.UDPAddr{IP: ip, Port: port})
		}
	}
	if len(addrs) != 0 {
		return addrs, nil
	}

	return nil, &net.DNSError{
		Err:        "lookup failed",
		Name:       hostname,
		IsNotFound: true,
	}
}
