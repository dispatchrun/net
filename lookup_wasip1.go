//go:build wasip1

package net

import (
	"fmt"
	"net"
	"net/netip"
	"strconv"

	"github.com/stealthrocket/net/syscall"
)

func lookupAddr(context, network, address string) (family int, addr syscall.Sockaddr, err error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return 0, nil, fmt.Errorf("not implemented: %s", network)
	}

	host, portstr, err := net.SplitHostPort(address)
	if err != nil {
		return 0, nil, err
	}
	port, err := lookupPort(portstr)
	if err != nil {
		return 0, nil, err
	}

	if host == "" {
		if context == "listen" {
			switch network {
			case "tcp", "tcp4":
				return syscall.AF_INET, &syscall.SockaddrInet4{Port: port}, nil
			case "tcp6":
				return syscall.AF_INET6, &syscall.SockaddrInet6{Port: port}, nil
			}
		}
		return 0, nil, fmt.Errorf("invalid address %q for %s", address, context)
	}
	ips, err := lookupIP(host)
	if err != nil {
		return 0, nil, err
	}
	if network == "tcp" || network == "tcp4" {
		for _, i := range ips {
			if inet4, ok := i.(*syscall.SockaddrInet4); ok {
				inet4.Port = port
				return syscall.AF_INET, inet4, nil
			}
		}
	}
	if network == "tcp" || network == "tcp6" {
		for _, i := range ips {
			if inet6, ok := i.(*syscall.SockaddrInet6); ok {
				inet6.Port = port
				return syscall.AF_INET6, inet6, nil
			}
		}
	}
	return 0, nil, fmt.Errorf("no route to host: %v", host)
}

func lookupIP(name string) ([]syscall.Sockaddr, error) {
	ip, err := netip.ParseAddr(name)
	if err == nil {
		if ip.Is4() {
			inet4 := &syscall.SockaddrInet4{Addr: ip.As4()}
			return []syscall.Sockaddr{inet4}, nil
		} else {
			inet6 := &syscall.SockaddrInet6{Addr: ip.As16()}
			return []syscall.Sockaddr{inet6}, nil
		}
	}
	return nil, fmt.Errorf("name resolution not implemented: %s", name)
}

func lookupPort(port string) (int, error) {
	// TODO: The Go stdlib consults /etc/services here first, allowing
	//  you to pass addresses like ":http"
	switch port {
	case "http":
		return 80, nil
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 0 || n > 65535 {
		return 0, fmt.Errorf("invalid port: %s", port)
	}
	return n, nil
}
