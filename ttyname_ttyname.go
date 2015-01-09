/*
	TTYNAME(3) in Go

	Copyright (C) 2014 Eric Lagergren

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

/* Written by Eric Lagergren */

package gnulib

import (
	"errors"
	"io"
	"os"
	"path"
	"syscall"
	"unsafe"
)

const dev = "/dev"

var (
	NotFound   = errors.New("device not found")
	NotTty     = errors.New("not a tty device")
	searchDevs = []string{
		"/dev/console",
		"/dev/wscons",
		"/dev/pts/",
		"/dev/vt/",
		"/dev/term/",
		"/dev/zcons/",
	}
	Stat = new(syscall.Stat_t)
)

func checkDirs(dir string) (*string, error) {
	var (
		rs      *string
		nameBuf = make([]byte, 256)
	)

	fi, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	dirBuf := make(DirentBuf)
	err = ReadDir(int(fi.Fd()), -1, &dirBuf)
	if err != nil && err != io.EOF {
		return nil, err
	}

	for _, v := range dirBuf {
		// quickly skip most entries
		if v.Ino != Stat.Ino {
			continue
		}

		_ = copy(nameBuf, int8ToByte(v.Name[:]))
		name := path.Join(dir, string(nameBuf[:clen(nameBuf)]))

		// Directories to skip
		if name == "/dev/stderr" ||
			name == "/dev/stdin" ||
			name == "/dev/stdout" ||
			len(name) >= 8 &&
				name[0:8] == "/dev/fd/" {
			continue
		}

		// We have to stat the file to determine its Rdev
		fstat := new(syscall.Stat_t)
		err = syscall.Stat(name, fstat)
		if err != nil {
			panic(err)
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
			return &name, nil
		}

	}

	return nil, NotFound
}

func isTty(fd uintptr) bool {
	var termios syscall.Termios

	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd,
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&termios)),
		0,
		0,
		0)
	return err == 0
}

func TtyName(fd uintptr) (*string, error) {
	var name *string

	if !isTty(fd) {
		return nil, NotTty
	}

	// gather inode and rdev info about fd
	err := syscall.Fstat(int(fd), Stat)
	if err != nil {
		return nil, err
	}

	// loop over most likely directories
	for _, v := range searchDevs {
		name, _ = checkDirs(v)
		if name != nil {
			return name, nil
		}
	}

	// if we can't find it do full scan of /dev/
	if name == nil {
		name, _ = checkDirs(dev)
		return name, nil
	}

	return nil, NotFound
}
