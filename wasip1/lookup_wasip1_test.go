//go:build wasip1

package wasip1

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestLookupAddr(t *testing.T) {
	tests := []struct {
		op      string
		network string
		address string
		addrs   []net.Addr
		err     error
	}{
		{
			op:      "dial",
			network: "tcp",
			address: "example.com:443",
			addrs: []net.Addr{
				&net.TCPAddr{IP: net.IPv4(10, 0, 0, 1), Port: 443},
				&net.TCPAddr{IP: net.IPv4(10, 0, 0, 2), Port: 443},
				&net.TCPAddr{IP: net.ParseIP("fe80::ec34:70ff:fe53:470e"), Port: 443},
			},
		},

		{
			op:      "dial",
			network: "tcp4",
			address: "example.com:443",
			addrs: []net.Addr{
				&net.TCPAddr{IP: net.IPv4(10, 0, 0, 1), Port: 443},
				&net.TCPAddr{IP: net.IPv4(10, 0, 0, 2), Port: 443},
			},
		},

		{
			op:      "dial",
			network: "tcp6",
			address: "example.com:443",
			addrs: []net.Addr{
				&net.TCPAddr{IP: net.ParseIP("fe80::ec34:70ff:fe53:470e"), Port: 443},
			},
		},

		{
			op:      "dial",
			network: "udp",
			address: "example.com:80",
			addrs: []net.Addr{
				&net.UDPAddr{IP: net.IPv4(10, 0, 0, 1), Port: 80},
				&net.UDPAddr{IP: net.IPv4(10, 0, 0, 2), Port: 80},
				&net.UDPAddr{IP: net.ParseIP("fe80::ec34:70ff:fe53:470e"), Port: 80},
			},
		},

		{
			op:      "dial",
			network: "udp4",
			address: "example.com:80",
			addrs: []net.Addr{
				&net.UDPAddr{IP: net.IPv4(10, 0, 0, 1), Port: 80},
				&net.UDPAddr{IP: net.IPv4(10, 0, 0, 2), Port: 80},
			},
		},

		{
			op:      "dial",
			network: "udp6",
			address: "example.com:80",
			addrs: []net.Addr{
				&net.UDPAddr{IP: net.ParseIP("fe80::ec34:70ff:fe53:470e"), Port: 80},
			},
		},

		{
			op:      "dial",
			network: "unix",
			address: "example.sock",
			addrs: []net.Addr{
				&net.UnixAddr{Net: "unix", Name: "example.sock"},
			},
		},

		{
			op:      "listen",
			network: "tcp",
			address: ":443",
			addrs: []net.Addr{
				&net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 443},
			},
		},

		{
			op:      "listen",
			network: "tcp4",
			address: ":443",
			addrs: []net.Addr{
				&net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 443},
			},
		},

		{
			op:      "listen",
			network: "tcp6",
			address: ":443",
			addrs: []net.Addr{
				&net.TCPAddr{IP: net.ParseIP("::"), Port: 443},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.op+" "+test.network+" "+test.address, func(t *testing.T) {
			addrs, err := lookupAddr(context.Background(), test.op, test.network, test.address)
			if !errors.Is(err, test.err) {
				t.Errorf("errors mismatch:\nwant = %v\ngot  = %v", test.err, err)
			}
			assertEqualAllAddrs(t, addrs, test.addrs)
		})
	}
}

func assertEqualAllAddrs(t *testing.T, addrs1, addrs2 []net.Addr) {
	if len(addrs1) != len(addrs2) {
		t.Errorf("number of addresses mismatch: %d != %d", len(addrs1), len(addrs2))
		t.Logf("   got: %v", addrs1)
		t.Logf("expect: %v", addrs2)
	} else {
		for i := range addrs1 {
			assertEqualAddr(t, addrs1[i], addrs2[i])
		}
	}
}

func assertEqualAddr(t *testing.T, addr1, addr2 net.Addr) {
	switch a1 := addr1.(type) {
	case *net.TCPAddr:
		if a2, ok := addr2.(*net.TCPAddr); ok {
			assertEqualTCPAddr(t, a1, a2)
			return
		}
	case *net.UDPAddr:
		if a2, ok := addr2.(*net.UDPAddr); ok {
			assertEqualUDPAddr(t, a1, a2)
			return
		}
	case *net.UnixAddr:
		if a2, ok := addr2.(*net.UnixAddr); ok {
			assertEqualUnixAddr(t, a1, a2)
			return
		}
	}
	t.Errorf("cannot compare addresses of type %T and %T", addr1, addr2)
}

func assertEqualTCPAddr(t *testing.T, addr1, addr2 *net.TCPAddr) {
	assertEqualIPAndPort(t,
		addr1.IP,
		addr2.IP,
		addr1.Port,
		addr2.Port,
		addr1.Zone,
		addr2.Zone,
	)
}

func assertEqualUDPAddr(t *testing.T, addr1, addr2 *net.UDPAddr) {
	assertEqualIPAndPort(t,
		addr1.IP,
		addr2.IP,
		addr1.Port,
		addr2.Port,
		addr1.Zone,
		addr2.Zone,
	)
}

func assertEqualUnixAddr(t *testing.T, addr1, addr2 *net.UnixAddr) {
	if addr1.Net != addr2.Net {
		t.Errorf("networks mismatch: %q != %q", addr1.Net, addr2.Net)
	}
	if addr1.Name != addr2.Name {
		t.Errorf("unix socket names mismatch: %q != %q", addr1.Name, addr2.Name)
	}
}

func assertEqualIPAndPort(t *testing.T, ip1, ip2 net.IP, port1, port2 int, zone1, zone2 string) {
	if !ip1.Equal(ip2) {
		t.Errorf("ip addreses mismatch: %v != %v", ip1, ip2)
	}
	if port1 != port2 {
		t.Errorf("ports mismatch: %d != %d", port1, port2)
	}
	if zone1 != zone2 {
		t.Errorf("zones mismatch: %q != %q", zone1, zone2)
	}
}
