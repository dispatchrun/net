// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syscall

import (
	"encoding/binary"
	"runtime"
	"unsafe"
)

const (
	AF_UNSPEC = iota
	AF_INET
	AF_INET6
	AF_UNIX
)

const (
	SOCK_ANY = iota
	SOCK_DGRAM
	SOCK_STREAM
)

const (
	SOL_SOCKET = iota
)

const (
	SO_REUSEADDR = iota
	_
	SO_ERROR
	_
	SO_BROADCAST
)

const (
	AI_PASSIVE = 1 << iota
	_
	AI_NUMERICHOST
	AI_NUMERICSERV
)

const (
	IPPROTO_IP = iota
	IPPROTO_TCP
	IPPROTO_UDP
)

type Sockaddr interface {
	sockaddr() (unsafe.Pointer, error)
	sockport() int
}

type SockaddrInet4 struct {
	Port int
	Addr [4]byte

	raw addressBuffer
}

func (s *SockaddrInet4) sockaddr() (unsafe.Pointer, error) {
	s.raw.bufLen = 4
	s.raw.buf = uintptr32(uintptr(unsafe.Pointer(&s.Addr)))
	return unsafe.Pointer(&s.raw), nil
}

func (s *SockaddrInet4) sockport() int {
	return s.Port
}

type SockaddrInet6 struct {
	Port   int
	ZoneId uint32
	Addr   [16]byte

	raw addressBuffer
}

func (s *SockaddrInet6) sockaddr() (unsafe.Pointer, error) {
	if s.ZoneId != 0 {
		return nil, ENOTSUP
	}
	s.raw.bufLen = 16
	s.raw.buf = uintptr32(uintptr(unsafe.Pointer(&s.Addr)))
	return unsafe.Pointer(&s.raw), nil
}

func (s *SockaddrInet6) sockport() int {
	return s.Port
}

type SockaddrUnix struct {
	Name string
}

func (s *SockaddrUnix) sockaddr() (unsafe.Pointer, error) {
	return nil, ENOSYS
}

func (s *SockaddrUnix) sockport() int {
	return 0
}

type uintptr32 = uint32
type size = uint32

type addressBuffer struct {
	buf    uintptr32
	bufLen size
}

type RawSockaddrAny struct {
	family uint16
	addr   [126]byte
}

//go:wasmimport wasi_snapshot_preview1 sock_open
func sock_open(af int32, socktype int32, fd unsafe.Pointer) Errno

//go:wasmimport wasi_snapshot_preview1 sock_bind
func sock_bind(fd int32, addr unsafe.Pointer, port uint32) Errno

//go:wasmimport wasi_snapshot_preview1 sock_listen
func sock_listen(fd int32, backlog int32) Errno

//go:wasmimport wasi_snapshot_preview1 sock_connect
func sock_connect(fd int32, addr unsafe.Pointer, port uint32) Errno

//go:wasmimport wasi_snapshot_preview1 sock_getsockopt
func sock_getsockopt(fd int32, level uint32, name uint32, value unsafe.Pointer, valueLen uint32) Errno

//go:wasmimport wasi_snapshot_preview1 sock_setsockopt
func sock_setsockopt(fd int32, level uint32, name uint32, value unsafe.Pointer, valueLen uint32) Errno

//go:wasmimport wasi_snapshot_preview1 sock_getlocaladdr
func sock_getlocaladdr(fd int32, addr unsafe.Pointer, port unsafe.Pointer) Errno

//go:wasmimport wasi_snapshot_preview1 sock_getpeeraddr
func sock_getpeeraddr(fd int32, addr unsafe.Pointer, port unsafe.Pointer) Errno

//go:wasmimport wasi_snapshot_preview1 sock_getaddrinfo
func sock_getaddrinfo(
	node unsafe.Pointer,
	nodeLen uint32,
	service unsafe.Pointer,
	serviceLen uint32,
	hints unsafe.Pointer,
	res unsafe.Pointer,
	maxResLen uint32,
	resLen unsafe.Pointer,
) uint32

func Socket(proto, sotype, unused int) (fd int, err error) {
	var newfd int32
	errno := sock_open(int32(proto), int32(sotype), unsafe.Pointer(&newfd))
	return int(newfd), errnoErr(errno)
}

func Bind(fd int, sa Sockaddr) error {
	rawaddr, err := sa.sockaddr()
	if err != nil {
		return err
	}
	errno := sock_bind(int32(fd), rawaddr, uint32(sa.sockport()))
	runtime.KeepAlive(sa)
	return errnoErr(errno)
}

func Listen(fd int, backlog int) error {
	errno := sock_listen(int32(fd), int32(backlog))
	return errnoErr(errno)
}

func Connect(fd int, sa Sockaddr) error {
	rawaddr, err := sa.sockaddr()
	if err != nil {
		return err
	}
	errno := sock_connect(int32(fd), rawaddr, uint32(sa.sockport()))
	runtime.KeepAlive(sa)
	return errnoErr(errno)
}

func GetsockoptInt(fd, level, opt int) (value int, err error) {
	var n int32
	errno := sock_getsockopt(int32(fd), uint32(level), uint32(opt), unsafe.Pointer(&n), 4)
	return int(n), errnoErr(errno)
}

func SetsockoptInt(fd, level, opt int, value int) error {
	var n = int32(value)
	errno := sock_setsockopt(int32(fd), uint32(level), uint32(opt), unsafe.Pointer(&n), 4)
	return errnoErr(errno)
}

func Getsockname(fd int) (sa Sockaddr, err error) {
	var rsa RawSockaddrAny
	buf := addressBuffer{
		buf:    uintptr32(uintptr(unsafe.Pointer(&rsa))),
		bufLen: uint32(unsafe.Sizeof(rsa)),
	}
	var port uint32
	errno := sock_getlocaladdr(int32(fd), unsafe.Pointer(&buf), unsafe.Pointer(&port))
	if errno != 0 {
		return nil, errnoErr(errno)
	}
	return anyToSockaddr(&rsa, int(port))
}

func Getpeername(fd int) (Sockaddr, error) {
	var rsa RawSockaddrAny
	buf := addressBuffer{
		buf:    uintptr32(uintptr(unsafe.Pointer(&rsa))),
		bufLen: uint32(unsafe.Sizeof(rsa)),
	}
	var port uint32
	errno := sock_getpeeraddr(int32(fd), unsafe.Pointer(&buf), unsafe.Pointer(&port))
	if errno != 0 {
		return nil, errnoErr(errno)
	}
	return anyToSockaddr(&rsa, int(port))
}

func anyToSockaddr(rsa *RawSockaddrAny, port int) (Sockaddr, error) {
	switch rsa.family {
	case AF_INET:
		addr := SockaddrInet4{Port: port}
		copy(addr.Addr[:], rsa.addr[:])
		return &addr, nil
	case AF_INET6:
		addr := SockaddrInet6{Port: port}
		copy(addr.Addr[:], rsa.addr[:])
		return &addr, nil
	default:
		return nil, ENOTSUP
	}
}

// https://github.com/WasmEdge/WasmEdge/blob/434e1fb4690/thirdparty/wasi/api.hpp#L1885
type addrInfo struct {
	ai_flags        uint16
	ai_family       uint8
	ai_socktype     uint8
	ai_protocol     uint32
	ai_addrlen      uint32
	ai_addr         uintptr32 // *sockAddr
	ai_canonname    uintptr32 // null-terminated string
	ai_canonnamelen uint32
	ai_next         uintptr32 // *addrInfo
}

type sockAddr struct {
	sa_family   uint32
	sa_data_len uint32
	sa_data     uintptr32
	_           [4]byte
}

type AddrInfo struct {
	Flags         int
	Family        int
	SocketType    int
	Protocol      int
	Address       Sockaddr
	CanonicalName string

	addrInfo
	sockAddr
	sockData  [26]byte
	cannoname [30]byte
	inet4addr SockaddrInet4
	inet6addr SockaddrInet6
}

func Getaddrinfo(name, service string, hints AddrInfo, results []AddrInfo) (int, error) {
	// For compatibility with WasmEdge, make sure strings are null-terminated.
	if len(name) > 0 && name[len(name)-1] != 0 {
		name = string(append([]byte(name), 0))
	}
	if len(service) > 0 && service[len(service)-1] != 0 {
		service = string(append([]byte(service), 0))
	}

	hints.addrInfo = addrInfo{
		ai_flags:    uint16(hints.Flags),
		ai_family:   uint8(hints.Family),
		ai_socktype: uint8(hints.SocketType),
		ai_protocol: uint32(hints.Protocol),
	}
	for i := range results {
		results[i].sockAddr = sockAddr{
			sa_family:   0,
			sa_data_len: uint32(unsafe.Sizeof(AddrInfo{}.sockData)),
			sa_data:     uintptr32(uintptr(unsafe.Pointer(&results[i].sockData))),
		}
		results[i].addrInfo = addrInfo{
			ai_flags:        0,
			ai_family:       0,
			ai_socktype:     0,
			ai_protocol:     0,
			ai_addrlen:      uint32(unsafe.Sizeof(sockAddr{})),
			ai_addr:         uintptr32(uintptr(unsafe.Pointer(&results[i].sockAddr))),
			ai_canonname:    uintptr32(uintptr(unsafe.Pointer(&results[i].cannoname))),
			ai_canonnamelen: uint32(unsafe.Sizeof(AddrInfo{}.cannoname)),
		}
		if i > 0 {
			results[i-1].addrInfo.ai_next = uintptr32(uintptr(unsafe.Pointer(&results[i-1].addrInfo)))
		}
	}

	resPtr := uintptr32(uintptr(unsafe.Pointer(&results[0].addrInfo)))

	var n uint32
	errno := sock_getaddrinfo(
		unsafe.Pointer(unsafe.StringData(name)),
		uint32(len(name)),
		unsafe.Pointer(unsafe.StringData(service)),
		uint32(len(service)),
		unsafe.Pointer(&hints.addrInfo),
		unsafe.Pointer(&resPtr),
		uint32(len(results)),
		unsafe.Pointer(&n),
	)
	if errno != 0 {
		return 0, errnoErr(Errno(errno))
	}

	for i := range results[:n] {
		r := &results[i]
		port := binary.BigEndian.Uint16(results[i].sockData[:2])
		switch results[i].sockAddr.sa_family {
		case AF_INET:
			r.inet4addr.Port = int(port)
			copy(r.inet4addr.Addr[:], results[i].sockData[2:])
			r.Address = &r.inet4addr
		case AF_INET6:
			r.inet6addr.Port = int(port)
			r.Address = &r.inet6addr
			copy(r.inet4addr.Addr[:], results[i].sockData[2:])
		default:
			r.Address = nil
		}
		// TODO: canonical names
	}
	return int(n), nil
}
