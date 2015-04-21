/*
	TTYNAME(3) in Go

	Copyright (C) 2015 Eric Lagergren

	This program is free software: you can redistribute it and/or modify
	it under the terms of the Lesser GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU Lesser General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

/* Written by Eric Lagergren */

package ttyname

import (
	"bytes"
	"errors"
	"syscall"

	"github.com/EricLagerg/go-gnulib/windows"
)

const delim = 0 // null byte

var (
	NotTty = errors.New("not a tty device")
)

func count(s []byte) int64 {
	count := int64(0)
	i := 0
	for i < len(s) {
		if s[i] != delim {
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
func TtyName(fd uintptr) (string, error) {

	// Does FD even describe a terminal? ;)
	if !IsAtty(fd) {
		return "", NotTty
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
