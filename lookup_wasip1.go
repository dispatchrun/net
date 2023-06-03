package net

import (
	"fmt"
	"net"
)

func lookupAddr(context, network, address string) (net.Addr, error) {
	switch network {
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
	case "unix", "unixgram":
		return &net.UnixAddr{Name: address, Net: network}, nil
	default:
		return nil, fmt.Errorf("not implemented: %s", network)
	}
	host, portstr, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	port, err := net.LookupPort(network, portstr)
	if err != nil {
		return nil, err
	}
	if host == "" {
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
	ips, err := net.LookupIP(host)
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
	return nil, fmt.Errorf("no route to host: %v", host)
}
