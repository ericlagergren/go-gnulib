// +build !windows

package ttyname

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

func IsAtty(fd uintptr) bool {
	var termios unix.Termios

	_, _, err := unix.Syscall6(unix.SYS_IOCTL, fd,
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&termios)),
		0,
		0,
		0)
	return err == 0
}
