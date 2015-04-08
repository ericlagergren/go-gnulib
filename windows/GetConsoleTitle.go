// +build windows

package k32

import (
	"syscall"
	"unsafe"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32")
	procGetConsoleTitleA = kernel32.NewProc("GetConsoleTitleA") // ANSI
	procGetConsoleTitleW = kernel32.NewProc("GetConsoleTitleW") // Unicode
)

func GetConsoleTitleA(buf []byte) error {
	size := len(buf)

	var p *byte
	if size > 0 {
		p = &buf[0]
	}

	length, _, err := syscall.Syscall(procGetConsoleTitleA.Addr(),
		2,
		uintptr(unsafe.Pointer(p)),
		uintptr(size), 0)

	if length <= 0 {
		return err
	}

	return nil
}

// Returns length as well so you can use
// utf16.decode(buf[0:length])
func GetConsoleTitleW(buf []uint16) (int, error) {
	size := len(buf)

	var p *uint16
	if size > 0 {
		p = &buf[0]
	}

	length, _, err := syscall.Syscall(procGetConsoleTitleW.Addr(),
		2,
		uintptr(unsafe.Pointer(p)),
		uintptr(size), 0)

	if length <= 0 {
		return int(length), err
	}

	return int(length), nil
}
