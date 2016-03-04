// +build s390x

// NOTE: DEPRECIATED

package posix

import (
	"syscall"
)

// NOTE: DEPRECIATED
// DONTNEED and NOREUSE differ on s390x
// http://lxr.free-electrons.com/source/include/uapi/linux/fadvise.h#L14
const (
	FADVISE_NORMAL     = 0x0
	FADVISE_RANDOM     = 0x1
	FADVISE_SEQUENTIAL = 0x2
	FADVISE_WILLNEED   = 0x3
	FADVISE_DONTNEED   = 0x6
	FADVISE_NOREUSE    = 0x7
)

// NOTE: DEPRECIATED
func Fadvise64(fd int, offset int64, length int64, advice int) error {
	_, _, errno := syscall.Syscall6(syscall.SYS_FADVISE64,
		uintptr(fd),
		uintptr(offset),
		uintptr(length),
		uintptr(advice), 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}
