package fts

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

func fstatat(fd int, path string, stat *unix.Stat_t, flags int) (err error) {
	var _p0 *byte
	_p0, err = unix.BytePtrFromString(path)
	if err != nil {
		return
	}

	_, _, e1 := unix.Syscall6(unix.SYS_NEWFSTATAT,
		uintptr(fd),
		uintptr(unsafe.Pointer(_p0)),
		uintptr(unsafe.Pointer(stat)),
		uintptr(flags),
		0,
		0)

	if e1 != 0 {
		err = e1
	}
	return
}
