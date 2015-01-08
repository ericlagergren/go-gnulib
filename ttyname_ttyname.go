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
	"os"
	"path"
	"syscall"
	"unsafe"
)

const (
	// The single letters are the abbreviations
	// used by the String method's formatting.
	ModeDir        = 1 << (32 - 1 - iota) // d: is a directory
	ModeAppend                            // a: append-only
	ModeExclusive                         // l: exclusive use
	ModeTemporary                         // T: temporary file (not backed up)
	ModeSymlink                           // L: symbolic link
	ModeDevice                            // D: device file
	ModeNamedPipe                         // p: named pipe (FIFO)
	ModeSocket                            // S: Unix domain socket
	ModeSetuid                            // u: setuid
	ModeSetgid                            // g: setgid
	ModeCharDevice                        // c: Unix character device, when ModeDevice is set
	ModeSticky                            // t: sticky

	// Mask for the type bits. For regular files, none will be set.
	ModeType = ModeDir | ModeSymlink | ModeNamedPipe | ModeSocket | ModeDevice

	ModePerm = 0777 // permission bits
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

func fileMode(longMode uint32) uint32 {
	mode := longMode & ModePerm
	switch longMode & syscall.S_IFMT {
	case syscall.S_IFBLK:
		mode |= ModeDevice
	case syscall.S_IFCHR:
		mode |= ModeDevice | ModeCharDevice
	case syscall.S_IFDIR:
		mode |= ModeDir
	case syscall.S_IFIFO:
		mode |= ModeNamedPipe
	case syscall.S_IFLNK:
		mode |= ModeSymlink
	case syscall.S_IFREG:
		// nothing to do
	case syscall.S_IFSOCK:
		mode |= ModeSocket
	}
	if longMode&syscall.S_ISGID != 0 {
		mode |= ModeSetgid
	}
	if longMode&syscall.S_ISUID != 0 {
		mode |= ModeSetuid
	}
	if longMode&syscall.S_ISVTX != 0 {
		mode |= ModeSticky
	}
	return mode
}

func checkDirs(dir string) (*string, error) {
	var (
		rs       *string
		fullPath string
	)

	fi, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	names, err := fi.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	for _, name := range names {
		fullPath = path.Join(dir, name)
		fstat := new(syscall.Stat_t)

		err = syscall.Stat(fullPath, fstat)
		if err != nil {
			continue
		}

		// Directories to skip
		if fullPath == "/dev/stderr" ||
			fullPath == "/dev/stdin" ||
			fullPath == "/dev/stdout" ||
			len(fullPath) >= 8 &&
				fullPath[0:8] == "/dev/fd/" {
			continue
		}

		fmode := fileMode(fstat.Mode)
		if fmode&ModeDir != 0 {
			rs, err = checkDirs(fullPath)
			if err != nil {
				continue
			}

			return rs, nil
		}

		if fmode&ModeCharDevice != 0 &&
			fstat.Ino == Stat.Ino &&
			fstat.Rdev == Stat.Rdev {
			return &fullPath, nil
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
	}

	// if we can't find it do full scan of /dev/
	if name == nil {
		name, _ = checkDirs(dev)
		return name, nil
	}

	return nil, NotFound
}
