// Package fdopen implements the functions found inside <fdopen*.c>
package fd

import (
	"golang.org/x/sys/unix"

	"github.com/EricLagergren/go-gnulib/cwd"
	"github.com/EricLagergren/go-gnulib/dirent"
)

func isExpected(err error) bool {
	return err == unix.ENOTDIR ||
		err == unix.ENOENT ||
		err == unix.EPERM ||
		err == unix.EACCES ||
		err == unix.ENOSYS ||
		err == unix.EOPNOTSUPP
}

// OpenDir opens a directory for the directory referred to by fd.
func OpenDir(fd int) (*dirent.Stream, error) {
	stream, err := openWithDup(fd, -1, nil)
	if isExpected(err) {
		saved := &cwd.CWD{}
		saved.Save()
		stream, err = openWithDup(fd, -1, saved)
	}
	return stream, err
}

func openWithDup(fd, oldfd int, savedcwd *cwd.CWD) (*dirent.Stream, error) {
	dupfd, err := unix.Dup(fd)
	if dupfd < 0 && err == unix.EMFILE {
		dupfd = oldfd
	}

	if err != nil {
		return nil, err
	}

	var (
		dir   *dirent.Stream
		saved error
	)

	if dupfd < fd-1 && dupfd != oldfd {
		dir, saved = openWithDup(fd, oldfd, savedcwd)
	} else {
		unix.Close(fd)
		dir, saved = cloneOpenDir(dupfd, savedcwd)

		if dir == nil {
			if fd1, err := unix.Dup(dupfd); fd1 != fd {
				panic(err)
			}
		}
	}

	if dupfd != oldfd {
		unix.Close(dupfd)
	}

	return dir, saved
}

func cloneOpenDir(fd int, savedcwd *cwd.CWD) (*dirent.Stream, error) {
	if err := unix.Fchdir(fd); err != nil {
		return nil, err
	}
	stream, err := dirent.Open(".", 0)
	if err != nil {
		return nil, err
	}
	return stream, savedcwd.Restore()
}
