//go:build wasip1 && !getaddrinfo

package wasip1

import (
	"context"
	"net"
)

func lookupAddr(ctx context.Context, op, network, address string) ([]net.Addr, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	case "udp", "udp4", "udp6":
	case "unix", "unixgram":
		return []net.Addr{&net.UnixAddr{Name: address, Net: network}}, nil
	default:
		return nil, net.UnknownNetworkError(network)
	}

	hostname, service, err := net.SplitHostPort(address)
	if err != nil {
		return nil, net.InvalidAddrError(address)
	}

	port, err := net.DefaultResolver.LookupPort(ctx, network, service)
	if err != nil {
		return nil, err
	}

	if hostname == "" {
		if op == "listen" {
			switch network {
			case "tcp", "tcp4":
				return []net.Addr{&net.TCPAddr{IP: net.IPv4zero, Port: port}}, nil
			case "tcp6":
				return []net.Addr{&net.TCPAddr{IP: net.IPv6zero, Port: port}}, nil
			}
		}
		return nil, net.InvalidAddrError(address)
	}

	ipAddrs, err := net.DefaultResolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		return nil, err
	}

	addrs := make([]net.Addr, 0, len(ipAddrs))
	switch network {
	case "tcp", "tcp4", "tcp6":
		for _, ipAddr := range ipAddrs {
			addrs = append(addrs, &net.TCPAddr{
				IP:   ipAddr.IP,
				Zone: ipAddr.Zone,
				Port: port,
			})
		}
	case "udp", "udp4", "udp6":
		for _, ipAddr := range ipAddrs {
			addrs = append(addrs, &net.UDPAddr{
				IP:   ipAddr.IP,
				Zone: ipAddr.Zone,
				Port: port,
			})
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
