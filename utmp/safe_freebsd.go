package utmp

import (
	"fmt"
	"os"
	"syscall"
)

// A wrapper around os.OpenFile() that locks the file after opening
// Will return an error if the file cannot be opened or the file cannot
// be locked
func SafeOpen(name string) (*os.File, *syscall.Flock_t, error) {

	fi, err := os.OpenFile(name, os.O_RDWR, os.ModeExclusive)
	if err != nil {
		return nil, nil, err
	}

	// Lock the file so we're responsible
	lk := syscall.Flock_t{
		Type:   syscall.F_WRLCK,    // set write lock
		Whence: 0,                  // SEEK_SET
		Start:  0,                  // beginning of file
		Len:    0,                  // until EOF
		Pid:    int32(os.Getpid()), // our PID
	}

	err = syscall.FcntlFlock(fi.Fd(), syscall.F_SETLKW, &lk)

	// If we can't lock the file error out to prevent corruption
	if err != nil {
		return nil, nil, err
	}

	return fi, &lk, nil
}

// Unlocks the file and then closes it. Returns an error if the file
// cannot be closed; unlocking errors are ignored.
func SafeClose(file *os.File, lk *syscall.Flock_t) error {
	Unlock(file, lk)
	return file.Close()
}

// Unlock an open file. Unlocking errors are ignored.
func Unlock(file *os.File, lk *syscall.Flock_t) error {
	if lk == nil || file == nil {
		return fmt.Errorf("file or lock are nil file: %s lk: %v", file.Name(), lk)
	}

	lk.Type = syscall.F_ULOCK
	return syscall.FcntlFlock(file.Fd(), syscall.F_SETLK, lk)
}
