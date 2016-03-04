// Package cwd implements GNU's save-cwd.c
package cwd

import (
	"os"

	"github.com/EricLagergren/go-gnulib/chdir"
	"github.com/EricLagergren/go-gnulib/util"
	"github.com/EricLagergren/go-gnulib/ifdef"

	"golang.org/x/sys/unix"
)

// CWD represents a saved working directory used with fchdir.
type CWD struct {
	Desc int // FD
	Name string
}

// Getcwd is guaranteed to always return a string with a length
// equal to len(buf) or the entire path if nil.
func Getcwd(buf []byte) string {
	if buf == nil {
		buf = make([]byte, 4096)
	}
	for err := error(unix.ERANGE); err != nil; {
		buf = make([]byte, len(buf)*2)
		_, err = unix.Getcwd(buf)
	}
	return string(buf[:util.Clen(buf)])
}

// Save records the location of the current working directory in
// the receiver so you can use the Restore method to switch back
// to that directory.
func (c *CWD) Save() {
	file, err := os.OpenFile(".", ifdef.O_SEARCH, 0)

	c.Desc = -1
	if err == nil {
		c.Desc = int(file.Fd())
	}

	if c.Desc < 0 {
		c.Name = Getcwd(nil)
	}

	unix.CloseOnExec(c.Desc)
}

// Restore switches back to the directory stored in the receiver.
func (c CWD) Restore() error {
	if 0 <= c.Desc {
		return unix.Fchdir(c.Desc)
	}
	return chdir.ChdirLong(c.Name)
}
