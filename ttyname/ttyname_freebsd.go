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
	dev        = C._PATH_DEV // Usually "/dev/"
	maxPathLen = C.MAXNAMLEN // Usually 255
	fiodgname  = C.FIODGNAME
)

var (
	nameBuf = make([]byte, len(dev)+maxPathLen)
)

type fiodgname_arg struct {
	len int
	buf []byte
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
	fgn := fiodgname_arg{len, buf}

	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd,
		uintptr(fiodgname),
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
	used := len(dev)
	copy(nameBuf, dev)
	if !FDevName(fd, nameBuf[used:], length-used) {
		return "", ErrNotFound
	}

	return string(nameBuf), nil
}
