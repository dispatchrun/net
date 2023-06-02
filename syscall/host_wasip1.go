//go:build wasip1

package syscall

import "unsafe"

type size = uint32
type fdflags = uint32
type rights = uint64
type filetype = uint8
type siflags = uint32
type riflags = uint32
type roflags = uint32
type sdflags = uint32

//go:wasmimport wasi_snapshot_preview1 sock_accept
func sock_accept(fd int32, flags fdflags, newfd unsafe.Pointer) Errno

//go:wasmimport wasi_snapshot_preview1 sock_send
func sock_send(fd int32, iovs unsafe.Pointer, iovsLen size, siflags siflags, nread unsafe.Pointer) Errno

//go:wasmimport wasi_snapshot_preview1 sock_recv
func sock_recv(fd int32, iovs unsafe.Pointer, iovsLen size, riflags riflags, nread unsafe.Pointer, roflags unsafe.Pointer) Errno

//go:wasmimport wasi_snapshot_preview1 sock_shutdown
func sock_shutdown(fd int32, flags sdflags) Errno

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

//go:wasmimport wasi_snapshot_preview1 fd_close
func fd_close(fd int32) Errno

//go:wasmimport wasi_snapshot_preview1 fd_fdstat_get
func fd_fdstat_get(fd int32, buf unsafe.Pointer) Errno

//go:wasmimport wasi_snapshot_preview1 fd_fdstat_set_flags
func fd_fdstat_set_flags(fd int32, flags fdflags) Errno
