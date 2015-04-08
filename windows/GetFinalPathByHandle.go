// +build windows

package k32

import (
	"syscall"
	"unsafe"
)

var (
	modkernel32                   = syscall.NewLazyDLL("kernel32.dll")
	procGetFinalPathNameByHandleA = modkernel32.NewProc("GetFinalPathNameByHandleA") // ANSI
	procGetFinalPathNameByHandleW = modkernel32.NewProc("GetFinalPathNameByHandleW") // Unicode
)

// ANSI
func GetFinalPathNameByHandleA(handle syscall.Handle, buf []byte, flags int) error {
	size := len(buf)

	var p *byte
	if size > 0 {
		p = &buf[0]
	}

	length, _, err := syscall.Syscall6(procGetFinalPathNameByHandleA.Addr(),
		4,
		uintptr(handle),
		uintptr(unsafe.Pointer(p)),
		uintptr(size),
		uintptr(flags), 0, 0)

	if length <= 0 {
		return err
	}

	return nil
}

// Unicode
// Use utf16.decode(buf[0:n]) to get the path name
func GetFinalPathNameByHandleW(handle syscall.Handle, buf []uint16, flags int) (int, error) {
	size := len(buf)

	var p *uint16
	if size > 0 {
		p = &buf[0]
	}

	length, _, err := syscall.Syscall6(procGetFinalPathNameByHandleW.Addr(),
		4,
		uintptr(handle),
		uintptr(unsafe.Pointer(p)),
		uintptr(size),
		uintptr(flags), 0, 0)

	if length <= 0 {
		return int(length), err
	}

	return int(length), nil
}
