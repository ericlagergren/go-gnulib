package ttyname

// #include <paths.h>
// #include <dirent.h>
// #include <sys/ioctl.h>
import "C"
import (
	"syscall"
	"unsafe"
)

const (
	dev          = C._PATH_DEV // Usually "/dev/"
	maxPathLen   = C.MAXNAMLEN // Usually 255
	IOCPARM_MASK = C.IOCPARM_MAX
	IOC_IN       = C.IOC_IN
)

var (
	nameBuf = make([]byte, len(dev)+maxPathLen)
)

type fiodgname_arg struct {
	len int
	buf []byte
}

func ioc(inout, group, num, len uint64) uint64 {
	return ((inout) | (((len) & IOCPARM_MASK) << 16) | ((group) << 8) | (num))
}

func iow(g, n uint64, t fiodgname_arg) uint64 {
	return ioc(IOC_IN, g, n, uint64(unsafe.Sizeof(t)))
}

// IsAtty maps to libc's isatty
func IsAtty(fd uintptr) bool {
	var termios syscall.Termios

	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd,
		uintptr(syscall.TIOCGETA),
		uintptr(unsafe.Pointer(&termios)),
		0,
		0,
		0)
	return err == 0
}

func FDevName(fd uintptr, buf []byte, len int) bool {
	var (
		fgn       = fiodgname_arg{len, buf}
		FOIDGNAME = iow('f', 120, fgn)
	)

	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd,
		uintptr(FOIDGNAME),
		uintptr(unsafe.Pointer(&fgn)),
		0,
		0,
		0)
	return err == 0
}

func TtyName(fd uintptr) (string, error) {
	if !IsAtty(fd) {
		return "", ErrNotTty
	}

	length := len(nameBuf)
	if !FDevName(fd, nameBuf, length) {
		return "", ErrNotFound
	}

	return string(nameBuf)
}
