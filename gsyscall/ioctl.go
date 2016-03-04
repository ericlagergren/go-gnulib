package gsyscall

import (
	"syscall"
	"unsafe"
)

// File descriptor, a request, and a memory address
func Ioctl(fd int, request int, argp *int) (err error) {
	_, _, e1 := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(request),
		uintptr(unsafe.Pointer(argp)))
	if e1 != 0 {
		err = ErrnoErr(e1)
	}
	return
}
