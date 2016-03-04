package ttyname

import (
	"bytes"
	"syscall"

	k32 "github.com/EricLagergren/go-gnulib/windows"
)

func count(s []byte) int64 {
	count := int64(0)
	i := 0
	for i < len(s) {
		if s[i] != 0 {
			o := bytes.IndexByte(s[i:], 0)
			if o < 0 {
				break
			}
			i += o
		}
		count++
		i++
	}
	return count
}

// Might not catch everything...
func IsAtty(fd uintptr) bool {
	handle := syscall.Handle(fd)

	var mode uint32
	err := syscall.GetConsoleMode(handle, &mode)
	if err != nil {
		return false
	}
	return true
}

// Errors differ from the Unix version because of how the windows kernel
// process call thing works. Console/terminal stuff is tricky on Windows.
func ttyname(fd uintptr) (string, error) {

	// Does FD even describe a terminal? ;)
	if !IsAtty(fd) {
		return "", ErrNotTty
	}

	buf := make([]byte, syscall.MAX_PATH)
	err := k32.GetConsoleTitleA(buf)
	if err != nil {
		return "", err
	}

	// All null bytes, so no name, but err == nil so it's still a tty, it just
	// doesn't have a name.
	if count(buf) == syscall.MAX_PATH {
		return "", nil
	}

	return string(buf), nil
}
