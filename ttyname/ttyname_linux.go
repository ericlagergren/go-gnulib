package ttyname

import (
	"os"
	"path/filepath"

	"github.com/EricLagergren/go-gnulib/dirent"
	"github.com/EricLagergren/go-gnulib/util"

	"golang.org/x/sys/unix"
)

const (
	dev  = "/dev/"
	proc = "/proc/self/fd/0"
)

var (
	searchDevs = []string{
		"/dev/pts/",
		"/dev/console",
		"/dev/wscons",
		"/dev/vt/",
		"/dev/term/",
		"/dev/zcons/",
	}
	stat = &unix.Stat_t{}
)

// recursively walk through the directory at path until the correct device
// is found. Directories in searchDevs are automatically skipped.
func checkDirs(path string) (string, error) {
	var (
		rs      string
		nameBuf = make([]byte, 256)
	)

	stream, err := dirent.Open(path)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	dirBuf := stream.ReadAll()

	for _, v := range dirBuf {
		// quickly skip most entries
		if v.Ino != stat.Ino {
			continue
		}

		_ = copy(nameBuf, util.Int8ToByte(v.Name[:]))
		name := filepath.Join(path, string(nameBuf[:util.Clen(nameBuf)]))

		// Directories to skip
		if name == "/dev/stderr" ||
			name == "/dev/stdin" ||
			name == "/dev/stdout" ||
			(len(name) >= 8 &&
				name[0:8] == "/dev/fd/") {
			continue
		}

		// We have to stat the file to determine its Rdev
		fstat := &unix.Stat_t{}
		err = unix.Stat(name, fstat)
		if err != nil {
			continue
		}

		// file mode sans permission bits
		if os.FileMode(fstat.Mode).IsDir() {
			rs, err = checkDirs(name)
			if err != nil {
				continue
			}

			return rs, nil
		}

		if isProperDevice(fstat, stat) {
			return name, nil
		}

	}

	return "", ErrNotFound
}

func isProperDevice(fstat, stat *unix.Stat_t) bool {
	return os.FileMode(fstat.Mode)&os.ModeCharDevice == 0 &&
		fstat.Ino == stat.Ino &&
		fstat.Dev == stat.Dev
}

// Returns a string from a uintptr describing a file descriptor
func ttyname(fd uintptr) (string, error) {
	var name string

	// Does `fd` even describe a terminal? ;)
	if !IsAtty(fd) {
		return "", ErrNotTty
	}

	// Gather inode and rdev info about fd
	err := unix.Fstat(int(fd), stat)
	if err != nil {
		return "", err
	}

	// Needs to be a character device
	if os.FileMode(stat.Mode)&os.ModeCharDevice != 0 {
		return "", ErrNotTty
	}

	// strace of GNU's tty stats the return of readlink(/proc/self/fd)
	// let's do that instead, and fall back on searching /dev/
	if ret, _ := os.Readlink(proc); ret != "" {
		fstat := &unix.Stat_t{}
		_ = unix.Stat(ret, fstat)

		if isProperDevice(fstat, stat) {
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
		return checkDirs("/dev/")
	}

	return "", ErrNotFound
}
