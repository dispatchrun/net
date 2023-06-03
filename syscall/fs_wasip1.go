// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syscall

import "syscall"

func Close(fd int) error {
	return syscall.Close(fd)
}

func SetNonblock(fd int, nonblocking bool) error {
	return syscall.SetNonblock(fd, nonblocking)
}
