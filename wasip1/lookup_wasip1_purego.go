//go:build wasip1 && purego

package wasip1

import (
	"fmt"
	"net"
)

func init() {
	net.DefaultResolver.Dial = DialContext
}

func lookupAddr(context, network, address string) (net.Addr, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	case "udp", "udp4", "udp6":
	case "unix", "unixgram":
		return &net.UnixAddr{Name: address, Net: network}, nil
	default:
		return nil, net.UnknownNetworkError(network)
	}
	hostname, service, err := net.SplitHostPort(address)
	if err != nil {
		return nil, net.InvalidAddrError(address)
	}
	port, err := net.LookupPort(network, service)
	if err != nil {
		return nil, err
	}
	if hostname == "" {
		if context == "listen" {
			switch network {
			case "tcp", "tcp4":
				return &net.TCPAddr{IP: net.IPv4zero, Port: port}, nil
			case "tcp6":
				return &net.TCPAddr{IP: net.IPv6zero, Port: port}, nil
			}
		}
		return nil, fmt.Errorf("invalid address %q for %s", address, context)
	}
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}
	if network == "tcp" || network == "tcp4" {
		for _, ip := range ips {
			if len(ip) == net.IPv4len {
				return &net.TCPAddr{IP: ip, Port: port}, nil
			}
		}
	}
	if network == "tcp" || network == "tcp6" {
		for _, ip := range ips {
			if len(ip) == net.IPv6len {
				return &net.TCPAddr{IP: ip, Port: port}, nil
			}
		}
	}
	if network == "udp" || network == "udp4" {
		for _, ip := range ips {
			if len(ip) == net.IPv4len {
				return &net.UDPAddr{IP: ip, Port: port}, nil
			}
		}
	}
	if network == "udp" || network == "udp6" {
		for _, ip := range ips {
			if len(ip) == net.IPv6len {
				return &net.UDPAddr{IP: ip, Port: port}, nil
			}
		}
	}
	return nil, &net.DNSError{
		Err:        "lookup failed",
		Name:       hostname,
		IsNotFound: true,
	}
}
