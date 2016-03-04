// Package dirent implements the functions found inside <dirent.h>,
// sans scandir, fdopendir, alphasort, and versionsort.
package dirent

import (
	"io"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"

	"github.com/EricLagergren/go-gnulib/util"
)

// Stream mimics C's DIR structure. It's a stream that can be read from
// using Read, and will provide pointers to unix.Dirent structures.
type Stream struct {
	fd   int
	buf  []byte // directory I/O
	bp   int
	file *os.File // only used for s.CloseDir
}

// Open returns a new stream from the given directory as well any
// errors that have occurred. It could fail if the given path is
// not a directory. The stream is initialized with a 4096 byte buffer if the size parameter is left empty.
func Open(path string, size ...int) (*Stream, error) {
	file, err := os.OpenFile(path, unix.O_RDONLY|unix.O_DIRECTORY, 0666)
	if err != nil {
		return nil, err
	}

	s := 4096
	if len(size) > 0 {
		s = size[0]
	}

	return &Stream{
		fd:   int(file.Fd()),
		buf:  make([]byte, s),
		bp:   0,
		file: file,
	}, nil
}

// Close closes the associated stream.
func (s *Stream) Close() error {
	if !s.exists() {
		return unix.EINVAL
	}
	return s.file.Close()
}

// Read returns one unix.Dirent from the Stream. It will return a
// pointer to a unix.Dirent structure if one is found; otherwise,
// it returns an error. The error will be io.EOF if and only if
// the end of the directory is reached. The end of the directory
// is considered "reached" when either unix.Getdents returns
// 0 *or* the unix.Dirent.Reclen == 0.
func (s *Stream) Read() (*unix.Dirent, error) {

	// Empty buffer, refill.
	if s.bp == 0 || s.bp >= len(s.buf) {
		n, err := unix.Getdents(s.fd, s.buf)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, io.EOF
		}
		s.bp = 0
	}

	dirent := *(*unix.Dirent)(unsafe.Pointer(&s.buf[s.bp]))
	if dirent.Reclen == 0 {
		return nil, io.EOF
	}

	s.bp += int(dirent.Reclen)

	if isAbsent(&dirent) {
		return s.Read()
	}

	b := (*[256]byte)(unsafe.Pointer(&dirent.Name[0]))
	bn := string(b[0:util.Clen(b[:])])
	if bn == "." || bn == ".." {
		return s.Read()
	}

	return &dirent, nil
}

// ReadAll is a simple helper function that returns the entire
// (unix.Dirent) contents of a directory.
func (s *Stream) ReadAll() []*unix.Dirent {
	var (
		buf []*unix.Dirent
		ent *unix.Dirent
		err error
	)

	for {
		if ent, err = s.Read(); ent == nil || err != nil {
			break
		}
		buf = append(buf, ent)
	}
	return buf
}

// Rewind resets the stream back to the beginning, similar to closing
// and re-opening a file.
func (s *Stream) Rewind() error {
	if !s.exists() {
		return unix.EINVAL
	}

	s.bp = 0
	_, err := s.file.Seek(0, os.SEEK_SET)
	return err
}

// Seek sets the location in the stream from which the next Read
// call will start.
func (s *Stream) Seek(loc int64) error {
	if !s.exists() {
		return unix.EINVAL
	}
	_, err := s.file.Seek(loc, os.SEEK_SET)
	return err
}

// Tell returns the current location in the directory stream.
// Note: Make *no* assumptions on the return value of this method.
// It is not guaranteed to return a simple offset, and blindly
// ignores errors, returning -1 if any errors are found.
func (s *Stream) Tell() int64 {
	if !s.exists() {
		return -1
	}

	cur, err := s.file.Seek(0, os.SEEK_CUR)
	if err != nil {
		cur = -1
	}
	return cur
}

// Fd returns the file descriptor for the given stream.
func (s *Stream) Fd() uintptr {
	if !s.exists() {
		return ^(uintptr(0))
	}
	return uintptr(s.fd)
}

// exists returns true if we can access the file (if any) inside the
// Stream pointer.
func (s *Stream) exists() bool { return s != nil && s.file != nil }
