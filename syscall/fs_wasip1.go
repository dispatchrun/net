// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build wasip1

package syscall

import "unsafe"

const (
	FDFLAG_NONBLOCK = 0x4
)

type uintptr32 = uint32

type iovec struct {
	buf    uintptr32
	bufLen size
}

func bytesPointer(b []byte) unsafe.Pointer {
	return unsafe.Pointer(unsafe.SliceData(b))
}

func makeIOVec(b []byte) unsafe.Pointer {
	return unsafe.Pointer(&iovec{
		buf:    uintptr32(uintptr(bytesPointer(b))),
		bufLen: size(len(b)),
	})
}

func Close(fd int) error {
	errno := fd_close(int32(fd))
	return errnoErr(errno)
}

type fdstat struct {
	filetype         filetype
	fdflags          uint16
	rightsBase       rights
	rightsInheriting rights
}

func fd_fdstat_get_flags(fd int) (uint32, error) {
	var stat fdstat
	errno := fd_fdstat_get(int32(fd), unsafe.Pointer(&stat))
	return uint32(stat.fdflags), errnoErr(errno)
}

func SetNonblock(fd int, nonblocking bool) error {
	flags, err := fd_fdstat_get_flags(fd)
	if err != nil {
		return err
	}
	if nonblocking {
		flags |= FDFLAG_NONBLOCK
	} else {
		flags &^= FDFLAG_NONBLOCK
	}
	errno := fd_fdstat_set_flags(int32(fd), flags)
	return errnoErr(errno)
}
