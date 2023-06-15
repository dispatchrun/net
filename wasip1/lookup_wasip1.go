//go:build wasip1 && !purego

package wasip1

import (
	"context"
	"errors"
	"net"
	"os"
	"strconv"
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
}

func lookupAddr(op, network, address string) (net.Addr, error) {
	var hints addrInfo
	switch network {
	case "tcp", "tcp4", "tcp6":
		hints.socketType = SOCK_STREAM
		hints.protocol = IPPROTO_TCP
	case "udp", "udp4", "udp6":
		hints.socketType = SOCK_DGRAM
		hints.protocol = IPPROTO_UDP
	case "unix", "unixgram":
		return &net.UnixAddr{Name: address, Net: network}, nil
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
		return nil, net.InvalidAddrError(address)
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

	results := make([]addrInfo, 16)
	n, err := getaddrinfo(hostname, service, &hints, results)
	if err != nil {
		addr := &netAddr{network, address}
		return nil, newOpError(op, addr, os.NewSyscallError("getaddrinfo", err))
	}
	results = results[:n]
	for _, r := range results {
		var ip net.IP
		var port int
		switch a := r.address.(type) {
		case *sockaddrInet4:
			ip = a.addr[:]
			port = a.port
		case *sockaddrInet6:
			ip = a.addr[:]
			port = a.port
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
