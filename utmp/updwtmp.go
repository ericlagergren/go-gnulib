package gnulib

import (
	"encoding/binary"
	"os"
	"syscall"
	"unsafe"
)

func UpdWtmp(fi *os.File, lk *syscall.Flock_t, u *Utmp) *UtmpError {
	su := unsafe.Sizeof(u)

	sz, err := fi.Seek(0, os.SEEK_END)
	if err != nil {
		// Cannot safely get file size in order to write
		Unlock(fi, lk)
		return &UtmpError{err, nil}
	}
	// If we can't write safely rewind the file and exit
	if sz%int64(su) != 0 {
		sz -= int64(su)
		err = syscall.Ftruncate(int(fi.Fd()), sz)
		if err != nil {
			Unlock(fi, lk)
			return &UtmpError{err, nil}
		}
	}

	if err != nil {
		Unlock(fi, lk)
		return &UtmpError{err, nil}
	}

	err = binary.Write(fi, binary.LittleEndian, &u)
	if err != nil {
		Unlock(fi, lk)
		return &UtmpError{err, nil}
	}

	return nil
}

func LogWtmp(fi *os.File, lk *syscall.Flock_t, line, user, host string) *UtmpError {
	u := new(Utmp)
	u.Time.GetTimeOfDay()
	u.Pid = int32(os.Getpid())
	_ = copy(u.Host[:], []byte(host))
	_ = copy(u.User[:], []byte(user))
	_ = copy(u.Line[:], []byte(line))

	err := UpdWtmp(fi, lk, u)
	if err != nil {
		return &UtmpError{err.WriteErr, err.LockErr}
	}

	return nil
}
