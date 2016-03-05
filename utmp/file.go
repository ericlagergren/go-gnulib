package utmp

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// File represents an *os.File with automatic locking
// on open and unlocking on close.
type File struct {
	lk *unix.Flock_t
	*os.File
}

var pid = int32(os.Getpid())

const (
	Reading = iota
	Writing
	Both
)

// Open is a wrapper around os.OpenFile() that locks the file after opening
// Returns an error if the file cannot be opened or the file cannot
// be locked.
func Open(name string, flags int) (*File, error) {
	file, err := os.OpenFile(name, os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		return nil, err
	}

	typ := int16(unix.F_RDLCK)
	if flags == Writing || flags == Both {
		typ = unix.F_WRLCK
	}

	// Lock the file so we're responsible
	lk := unix.Flock_t{
		Type: typ,
		Pid:  pid,
	}

	err = unix.FcntlFlock(file.Fd(), unix.F_SETLKW, &lk)
	if err != nil {
		return nil, err
	}
	return &File{
		lk:   &lk,
		File: file,
	}, nil
}

// Close unlocks the file and then closes it. Returns an error if the file
// cannot be closed; unlocking errors are ignored.
func (f *File) Close() error {
	f.unlock()
	return f.File.Close()
}

// unlock unlocks an open file. unlocking errors are ignored.
func (f *File) unlock() error {
	if f == nil {
		return errors.New("cannot unlock nil file")
	}
	if f.lk == nil {
		return errors.New("cannot unlock file with nil lock")
	}
	f.lk.Type = unix.F_UNLCK
	return unix.FcntlFlock(f.Fd(), unix.F_SETLK, f.lk)
}
