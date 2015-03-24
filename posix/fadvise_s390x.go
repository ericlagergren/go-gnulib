// +build s390x
package posix

import (
	"syscall"
)

const (
	FADVISE_NORMAL     = 0x0
	FADVISE_RANDOM     = 0x1
	FADVISE_SEQUENTIAL = 0x2
	FADVISE_WILLNEED   = 0x3
	FADVISE_DONTNEED   = 0x6
	FADVISE_NOREUSE    = 0x7
)

func Fadvise64(fd int, off int, length int, advice uint32) error {
	_, _, errno := syscall.Syscall6(syscall.SYS_FADVISE64,
		uintptr(fd),
		uintptr(off),
		uintptr(length),
		uintptr(advice), 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}
