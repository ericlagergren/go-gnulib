// +build arm

// NOTE: DEPRECIATED

package posix

import (
	"syscall"
)

// NOTE: DEPRECIATED
const (
	FADVISE_NORMAL     = 0x0
	FADVISE_RANDOM     = 0x1
	FADVISE_SEQUENTIAL = 0x2
	FADVISE_WILLNEED   = 0x3
	FADVISE_DONTNEED   = 0x4
	FADVISE_NOREUSE    = 0x5
)

// NOTE: DEPRECIATED
// ARM64 requires arguments to be realigned.
// http://linux.die.net/man/2/posix_fadvise
func Fadvise64(fd int, advice int64, offset int64, length int) error {
	_, _, errno := syscall.Syscall6(syscall.SYS_FADVISE64,
		uintptr(fd),
		uintptr(advice),
		uintptr(offset),
		uintptr(length), 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}
