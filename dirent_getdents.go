package gnulib

import (
	"fmt"
	"io"
	"syscall"
	"unsafe"
)

type DirentBuf map[int64]*syscall.Dirent

// Because we don't import os and dirInfo isn't exported anyway
type dirInfo struct {
	buf  []byte // buffer for directory I/O
	nbuf int    // length of buf; return value from Getdirentries
	bufp int    // location of next record in buf.
}

// Convert Dirent.Name to byte slice
func int8ToByte(rel [256]int8) []byte {
	s := [256]byte{}
	for i := 0; i < len(rel); i++ {
		s[i] = byte(rel[i])
	}
	return s[:clen8(rel)]
}

// Find length of a C-style string
func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}

// clen() but for int8 instead of []byte
func clen8(n [256]int8) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}

// change -1 to 0
func fixCount(n int, err error) (int, error) {
	if n < 0 {
		n = 0
	}
	return n, err
}

// Read a buffer (dir) into the pointer to the DirentBuf, `db`
// Returns remaining buffer size and number of entries created
// It's a modified version of syscall.ParseDirent in Go's standard library
func ParseDir(buf []byte, max int64, db *DirentBuf) (int, int) {
	orig := len(buf)

	i := int64(0)
	for max != 0 && len(buf) > 0 {
		dirent := (*syscall.Dirent)(unsafe.Pointer(&buf[0]))

		buf = buf[dirent.Reclen:]
		if dirent.Ino == 0 {
			continue
		}

		b := (*[257]byte)(unsafe.Pointer(&dirent.Name[0]))
		bn := string(b[0:clen(b[:])])
		if bn == "." || bn == ".." {
			continue
		}

		(*db)[i] = dirent
		max--
		i++
	}
	return orig - len(buf), int(i)
}

// Read through the directory calling ParseDir() to write to the DirentBuf
// specified by 'db'
// Returns error EOF when the directory has been walked; returns "readdirent"
// error if unable to perform a ReadDirent() syscall
// It's a modified version of (os) readdirnames() in Go's standard library
func ReadDir(fd int, n int64, db *DirentBuf) error {
	d := new(dirInfo)
	d.buf = make([]byte, 4096)

	size := n
	if size <= 0 {
		size = 100
		n = -1
	}

	for n != 0 {
		// Refill the buffer if necessary
		if d.bufp >= d.nbuf {
			d.bufp = 0
			var errno error
			d.nbuf, errno = fixCount(syscall.ReadDirent(fd, d.buf))
			if errno != nil {
				return os.NewSyscallError("readdirent", errno)
			}
			if d.nbuf <= 0 {
				break // EOF
			}
		}

		// Drain the buffer
		var nb, nc int
		nb, nc = ParseDir(d.buf[d.bufp:d.nbuf], n, db)
		d.bufp += nb
		n -= int64(nc)
	}
	if n >= 0 {
		return io.EOF
	}
	return nil
}

// Example usage
/*
func main() {
	fi, err := os.Open("/test/directory/")
	if err != nil {
		// handle err
	}
	defer fi.Close()

	db := make(DirentBuf)
	err = ReadDir(int(fi.Fd()), -1, &db)
	if err != nil && err != io.EOF {
		// handle err
	}

	if

	for _, v := range db {
		n := int8ToByte(v.Name)
		fmt.Println(string(n))
	}
}*/
