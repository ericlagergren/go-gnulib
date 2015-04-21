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
	"errors"
	"io"
	"os"
	"path"
	"syscall"
	"unsafe"

	"github.com/EricLagerg/go-gnulib/dirent"
	"github.com/EricLagerg/go-gnulib/general"
)

const (
	dev  = "/dev"
	proc = "/proc/self/fd/0"
)

var (
	NotFound   = errors.New("device not found")
	NotTty     = errors.New("not a tty device")
	searchDevs = []string{
		"/dev/pts/",
		"/dev/console",
		"/dev/wscons",
		"/dev/vt/",
		"/dev/term/",
		"/dev/zcons/",
	}
	Stat = new(syscall.Stat_t)
)

// recursively walk through the named directory `dir` until the correct device
// is found.
// Directories in []searchDevs are automatically skipped
func checkDirs(dir string) (string, error) {
	var (
		rs      string
		nameBuf = make([]byte, 256)
		dirBuf  = make([]syscall.Dirent, 256)
	)

	fi, err := os.Open(dir)
	if err != nil {
		return "", err
	}
	defer fi.Close()

	err = dirent.ReadDir(int(fi.Fd()), -1, &dirBuf)
	if err != nil && err != io.EOF {
		return "", err
	}

	for _, v := range dirBuf {
		// quickly skip most entries
		if v.Ino != Stat.Ino {
			continue
		}

		_ = copy(nameBuf, general.Int8ToByte(v.Name[:]))
		name := path.Join(dir, string(nameBuf[:general.Clen(nameBuf)]))

		// Directories to skip
		if name == "/dev/stderr" ||
			name == "/dev/stdin" ||
			name == "/dev/stdout" ||
			(len(name) >= 8 &&
				name[0:8] == "/dev/fd/") {
			continue
		}

		// We have to stat the file to determine its Rdev
		fstat := new(syscall.Stat_t)
		err = syscall.Stat(name, fstat)
		if err != nil {
			continue
		}

		// file mode sans permission bits
		fmode := os.FileMode(fstat.Mode)
		if fmode.IsDir() {
			rs, err = checkDirs(name)
			if err != nil {
				continue
			}

			return rs, nil
		}

		if fmode&os.ModeCharDevice == 0 &&
			fstat.Ino == Stat.Ino &&
			fstat.Rdev == Stat.Rdev {
			return name, nil
		}

	}

	return "", NotFound
}

// quick IsAtty check
func IsAtty(fd uintptr) bool {
	var termios syscall.Termios

	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd,
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&termios)),
		0,
		0,
		0)
	return err == 0
}

// Returns a string from a uintptr describing a file descriptor
func TtyName(fd uintptr) (string, error) {
	var name string

	// Does `fd` even describe a terminal? ;)
	if !IsAtty(fd) {
		return "", NotTty
	}

	// Gather inode and rdev info about fd
	err := syscall.Fstat(int(fd), Stat)
	if err != nil {
		return "", err
	}

	// Needs to be a character device
	if os.FileMode(Stat.Mode)&os.ModeCharDevice != 0 {
		return "", NotTty
	}

	// strace of GNU's tty stats the return of readlink(/proc/self/fd)
	// let's do that instead, and fall back on searching /dev/
	if ret, _ := os.Readlink(proc); ret != "" {
		fstat := new(syscall.Stat_t)
		_ = syscall.Stat(ret, fstat)

		if os.FileMode(fstat.Mode)&os.ModeCharDevice == 0 &&
			fstat.Ino == Stat.Ino &&
			fstat.Rdev == Stat.Rdev {
			return ret, nil
		}
	}

	// Loop over most likely directories second
	for _, v := range searchDevs {
		name, _ = checkDirs(v)
		if name != "" {
			return name, nil
		}
	}

	// If we can't find it above, do full scan of /dev/
	if name == "" {
		return checkDirs(dev)
	}

	return "", NotFound
}
