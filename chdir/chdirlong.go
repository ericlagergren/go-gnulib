// Package chdir implements GNU's chdir_long function.
package chdir

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/EricLagergren/go-gnulib/ifdef"
	"golang.org/x/sys/unix"
)

type cd struct{ fd int }

// ChdirLong should be used when you have a path that's > unix.PathMax.
// It breaks up the path into manageable sections and iterates through
// them, using Chdir on each section to simulate a normal
// Chdir. It'll return unix.ENAMETOOLONG if any section of the
// path (the parts between the separator) are > unix.PathMax.
func ChdirLong(name string) error {
	if err := unix.Chdir(name); err != nil || err == unix.ENAMETOOLONG {
		return err
	}

	parts := strings.Split(name, string(os.PathSeparator))

	var (
		total, prev, i int

		c = &cd{fd: unix.AT_FDCWD}
	)

	for i < len(parts) {
		for prev = i; i < len(parts); i++ {
			if total+len(parts[i]) > unix.PathMax {
				if len(parts[i]) > unix.PathMax {
					return unix.ENAMETOOLONG
				}
				break
			}
			total += len(parts[i]) + 1 // for the path separator
		}
		c.advance(join(prev == 0, parts[prev:i]...))
		total = 0
	}

	return unix.Fchdir(c.fd)
}

func join(begin bool, parts ...string) string {
	path := filepath.Join(parts...)
	if begin {
		path = string(os.PathSeparator) + path
	}
	return path
}

// advance moves up a directory at a time
func (c *cd) advance(dir string) error {
	newfd, err := unix.Openat(c.fd, dir,
		ifdef.O_SEARCH|
			unix.O_DIRECTORY|
			unix.O_NOCTTY|
			unix.O_NONBLOCK, 0)

	if err != nil {
		return err
	}

	c.close()
	c.fd = newfd
	return nil
}

func (c *cd) close() {
	if 0 <= c.fd {
		if err := unix.Close(c.fd); err != nil {
			panic(err)
		}
	}
}
