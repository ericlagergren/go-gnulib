package dirent

import (
	"io"
	"os"
	"syscall"
	"unsafe"

	"github.com/EricLagerg/go-gnulib/general"
)

type DirentBuf map[int64]*syscall.Dirent

// Because os' dirInfo isn't exported
type dirInfo struct {
	buf  []byte // buffer for directory I/O
	nbuf int    // length of buf; return value from Getdirentries
	bufp int    // location of next record in buf.
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
		bn := string(b[0:general.Clen(b[:])])
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
			d.nbuf, errno = general.FixCount(syscall.ReadDirent(fd, d.buf))
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
		n := Int8toByte(v.Name)
		fmt.Println(string(n))
	}
}*/
