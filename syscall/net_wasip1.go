// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syscall

import (
	"runtime"
	"unsafe"
)

const (
	_ = iota
	AF_INET
	AF_INET6
)

const (
	_ = iota
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
)

type uintptr32 = uint32

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

type addressBuffer struct {
	buf    uintptr32
	bufLen size
}

type RawSockaddrAny struct {
	family uint16
	addr   [126]byte
}

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
