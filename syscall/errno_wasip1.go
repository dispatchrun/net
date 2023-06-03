// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syscall

import (
	"syscall"
)

type Errno = syscall.Errno

const (
	EAGAIN      Errno = 6
	EINPROGRESS Errno = 26
	EINTR       Errno = 27
	EISCONN     Errno = 30
	ENOSYS      Errno = 52
	ENOTSUP     Errno = 58
)

// Do the interface allocations only once for common
// Errno values.
var errEAGAIN error = EAGAIN

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
//
// We set both noinline and nosplit to reduce code size, this function has many
// call sites in the syscall package, inlining it causes a significant increase
// of the compiled code; the function call ultimately does not make a difference
// in the performance of syscall functions since the time is dominated by calls
// to the imports and path resolution.
//
//go:noinline
//go:nosplit
func errnoErr(e Errno) error {
	switch e {
	case 0:
		return nil
	case EAGAIN:
		return errEAGAIN
	}
	return e
}
