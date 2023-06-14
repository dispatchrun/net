//go:build wasip1

package wasip1

// This file contains the definition of host imports compatible with the socket
// extensions from wasmedge v0.12+.

import (
	"encoding/binary"
	"runtime"
	"strings"
	"syscall"
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

type sockaddr interface {
	sockaddr() (unsafe.Pointer, error)
	sockport() int
}

type sockaddrInet4 struct {
	port int
	addr [4]byte
	raw  addressBuffer
}

func (s *sockaddrInet4) sockaddr() (unsafe.Pointer, error) {
	s.raw.bufLen = 4
	s.raw.buf = uintptr32(uintptr(unsafe.Pointer(&s.addr)))
	return unsafe.Pointer(&s.raw), nil
}

func (s *sockaddrInet4) sockport() int {
	return s.port
}

type sockaddrInet6 struct {
	port int
	zone uint32
	addr [16]byte
	raw  addressBuffer
}

func (s *sockaddrInet6) sockaddr() (unsafe.Pointer, error) {
	if s.zone != 0 {
		return nil, syscall.ENOTSUP
	}
	s.raw.bufLen = 16
	s.raw.buf = uintptr32(uintptr(unsafe.Pointer(&s.addr)))
	return unsafe.Pointer(&s.raw), nil
}

func (s *sockaddrInet6) sockport() int {
	return s.port
}

type sockaddrUnix struct {
	name string

	raw rawSockaddrAny
	buf addressBuffer
}

func (s *sockaddrUnix) sockaddr() (unsafe.Pointer, error) {
	s.raw.family = AF_UNIX
	if len(s.name) >= len(s.raw.addr)-1 {
		return nil, syscall.EINVAL
	}
	copy(s.raw.addr[:], s.name)
	s.raw.addr[len(s.name)] = 0
	s.buf.bufLen = 128
	s.buf.buf = uintptr32(uintptr(unsafe.Pointer(&s.raw)))
	return unsafe.Pointer(&s.buf), nil
}

func (s *sockaddrUnix) sockport() int {
	return 0
}

type uintptr32 = uint32
type size = uint32

type addressBuffer struct {
	buf    uintptr32
	bufLen size
}

type rawSockaddrAny struct {
	family uint16
	addr   [126]byte
}

//go:wasmimport wasi_snapshot_preview1 sock_open
//go:noescape
func sock_open(af int32, socktype int32, fd unsafe.Pointer) syscall.Errno

//go:wasmimport wasi_snapshot_preview1 sock_bind
//go:noescape
func sock_bind(fd int32, addr unsafe.Pointer, port uint32) syscall.Errno

//go:wasmimport wasi_snapshot_preview1 sock_listen
//go:noescape
func sock_listen(fd int32, backlog int32) syscall.Errno

//go:wasmimport wasi_snapshot_preview1 sock_connect
//go:noescape
func sock_connect(fd int32, addr unsafe.Pointer, port uint32) syscall.Errno

//go:wasmimport wasi_snapshot_preview1 sock_getsockopt
//go:noescape
func sock_getsockopt(fd int32, level uint32, name uint32, value unsafe.Pointer, valueLen uint32) syscall.Errno

//go:wasmimport wasi_snapshot_preview1 sock_setsockopt
//go:noescape
func sock_setsockopt(fd int32, level uint32, name uint32, value unsafe.Pointer, valueLen uint32) syscall.Errno

//go:wasmimport wasi_snapshot_preview1 sock_getlocaladdr
//go:noescape
func sock_getlocaladdr(fd int32, addr unsafe.Pointer, port unsafe.Pointer) syscall.Errno

//go:wasmimport wasi_snapshot_preview1 sock_getpeeraddr
//go:noescape
func sock_getpeeraddr(fd int32, addr unsafe.Pointer, port unsafe.Pointer) syscall.Errno

//go:wasmimport wasi_snapshot_preview1 sock_getaddrinfo
//go:noescape
func sock_getaddrinfo(
	node unsafe.Pointer,
	nodeLen uint32,
	service unsafe.Pointer,
	serviceLen uint32,
	hints unsafe.Pointer,
	res unsafe.Pointer,
	maxResLen uint32,
	resLen unsafe.Pointer,
) syscall.Errno

func socket(proto, sotype, unused int) (fd int, err error) {
	var newfd int32
	errno := sock_open(int32(proto), int32(sotype), unsafe.Pointer(&newfd))
	if errno != 0 {
		return -1, errno
	}
	return int(newfd), nil
}

func bind(fd int, sa sockaddr) error {
	rawaddr, err := sa.sockaddr()
	if err != nil {
		return err
	}
	errno := sock_bind(int32(fd), rawaddr, uint32(sa.sockport()))
	runtime.KeepAlive(sa)
	if errno != 0 {
		return errno
	}
	return nil
}

func listen(fd int, backlog int) error {
	if errno := sock_listen(int32(fd), int32(backlog)); errno != 0 {
		return errno
	}
	return nil
}

func connect(fd int, sa sockaddr) error {
	rawaddr, err := sa.sockaddr()
	if err != nil {
		return err
	}
	errno := sock_connect(int32(fd), rawaddr, uint32(sa.sockport()))
	runtime.KeepAlive(sa)
	if errno != 0 {
		return errno
	}
	return nil
}

func getsockopt(fd, level, opt int) (value int, err error) {
	var n int32
	errno := sock_getsockopt(int32(fd), uint32(level), uint32(opt), unsafe.Pointer(&n), 4)
	if errno != 0 {
		return 0, errno
	}
	return int(n), nil
}

func setsockopt(fd, level, opt int, value int) error {
	var n = int32(value)
	errno := sock_setsockopt(int32(fd), uint32(level), uint32(opt), unsafe.Pointer(&n), 4)
	if errno != 0 {
		return errno
	}
	return nil
}

func getsockname(fd int) (sa sockaddr, err error) {
	var rsa rawSockaddrAny
	buf := addressBuffer{
		buf:    uintptr32(uintptr(unsafe.Pointer(&rsa))),
		bufLen: uint32(unsafe.Sizeof(rsa)),
	}
	var port uint32
	errno := sock_getlocaladdr(int32(fd), unsafe.Pointer(&buf), unsafe.Pointer(&port))
	if errno != 0 {
		return nil, errno
	}
	return anyToSockaddr(&rsa, int(port))
}

func getpeername(fd int) (sockaddr, error) {
	var rsa rawSockaddrAny
	buf := addressBuffer{
		buf:    uintptr32(uintptr(unsafe.Pointer(&rsa))),
		bufLen: uint32(unsafe.Sizeof(rsa)),
	}
	var port uint32
	errno := sock_getpeeraddr(int32(fd), unsafe.Pointer(&buf), unsafe.Pointer(&port))
	if errno != 0 {
		return nil, errno
	}
	return anyToSockaddr(&rsa, int(port))
}

func anyToSockaddr(rsa *rawSockaddrAny, port int) (sockaddr, error) {
	switch rsa.family {
	case AF_INET:
		addr := sockaddrInet4{port: port}
		copy(addr.addr[:], rsa.addr[:])
		return &addr, nil
	case AF_INET6:
		addr := sockaddrInet6{port: port}
		copy(addr.addr[:], rsa.addr[:])
		return &addr, nil
	case AF_UNIX:
		addr := sockaddrUnix{}
		n := 0
		for ; n > len(rsa.addr); n++ {
			if rsa.addr[n] == 0 {
				break
			}
		}
		addr.name = string(rsa.addr[:n])
		return &addr, nil
	default:
		return nil, syscall.ENOTSUP
	}
}

// https://github.com/WasmEdge/WasmEdge/blob/434e1fb4690/thirdparty/wasi/api.hpp#L1885
type sockAddrInfo struct {
	ai_flags        uint16
	ai_family       uint8
	ai_socktype     uint8
	ai_protocol     uint32
	ai_addrlen      uint32
	ai_addr         uintptr32 // *sockAddr
	ai_canonname    uintptr32 // null-terminated string
	ai_canonnamelen uint32
	ai_next         uintptr32 // *sockAddrInfo
}

type sockAddr struct {
	sa_family   uint32
	sa_data_len uint32
	sa_data     uintptr32
	_           [4]byte
}

type addrInfo struct {
	flags      int
	family     int
	socketType int
	protocol   int
	address    sockaddr
	// canonicalName string

	sockAddrInfo
	sockAddr
	sockData  [26]byte
	cannoname [30]byte
	inet4addr sockaddrInet4
	inet6addr sockaddrInet6
}

func getaddrinfo(name, service string, hints *addrInfo, results []addrInfo) (int, error) {
	hints.sockAddrInfo = sockAddrInfo{
		ai_flags:    uint16(hints.flags),
		ai_family:   uint8(hints.family),
		ai_socktype: uint8(hints.socketType),
		ai_protocol: uint32(hints.protocol),
	}
	for i := range results {
		results[i].sockAddr = sockAddr{
			sa_family:   0,
			sa_data_len: uint32(unsafe.Sizeof(addrInfo{}.sockData)),
			sa_data:     uintptr32(uintptr(unsafe.Pointer(&results[i].sockData))),
		}
		results[i].sockAddrInfo = sockAddrInfo{
			ai_flags:        0,
			ai_family:       0,
			ai_socktype:     0,
			ai_protocol:     0,
			ai_addrlen:      uint32(unsafe.Sizeof(sockAddr{})),
			ai_addr:         uintptr32(uintptr(unsafe.Pointer(&results[i].sockAddr))),
			ai_canonname:    uintptr32(uintptr(unsafe.Pointer(&results[i].cannoname))),
			ai_canonnamelen: uint32(unsafe.Sizeof(addrInfo{}.cannoname)),
		}
		if i > 0 {
			results[i-1].sockAddrInfo.ai_next = uintptr32(uintptr(unsafe.Pointer(&results[i-1].sockAddrInfo)))
		}
	}

	resPtr := uintptr32(uintptr(unsafe.Pointer(&results[0].sockAddrInfo)))
	// For compatibility with WasmEdge, make sure strings are null-terminated.
	namePtr, nameLen := nullTerminatedString(name)
	servPtr, servLen := nullTerminatedString(service)

	var n uint32
	errno := sock_getaddrinfo(
		unsafe.Pointer(namePtr),
		uint32(nameLen),
		unsafe.Pointer(servPtr),
		uint32(servLen),
		unsafe.Pointer(&hints.sockAddrInfo),
		unsafe.Pointer(&resPtr),
		uint32(len(results)),
		unsafe.Pointer(&n),
	)
	if errno != 0 {
		return 0, errno
	}

	for i := range results[:n] {
		r := &results[i]
		port := binary.BigEndian.Uint16(results[i].sockData[:2])
		switch results[i].sockAddr.sa_family {
		case AF_INET:
			r.inet4addr.port = int(port)
			copy(r.inet4addr.addr[:], results[i].sockData[2:])
			r.address = &r.inet4addr
		case AF_INET6:
			r.inet6addr.port = int(port)
			r.address = &r.inet6addr
			copy(r.inet4addr.addr[:], results[i].sockData[2:])
		default:
			r.address = nil
		}
		// TODO: canonical names
	}
	return int(n), nil
}

func nullTerminatedString(s string) (*byte, int) {
	if n := strings.IndexByte(s, 0); n >= 0 {
		s = s[:n+1]
		return unsafe.StringData(s), len(s)
	} else {
		b := append([]byte(s), 0)
		return unsafe.SliceData(b), len(b)
	}
}
