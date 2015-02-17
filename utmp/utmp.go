/*
	GNU's utmp.c written in Go

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

/* Written by Eric Lagergren <ericscottlagergren@gmail.com> */

package utmp

import (
	"encoding/binary"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"

	"github.com/EricLagerg/go-gnulib/general"
)

// WriteErr is an error reading or writing to a file
// LockErr is an error locking or unlocking a file
type UtmpError struct {
	WriteErr error
	LockErr  error
}

// Similar to glibc's gettimeofday()
func (t *timeVal) GetTimeOfDay() {
	now := time.Now().Unix()
	t.Usec = int32(now / 1000)
	t.Sec = int32(now)
}

// A wrapper around os.OpenFile() that locks the file after opening
// Returns a pointer to the open fd, the lock struct, and an error/nil
func SafeOpen(name string, flag int, perm os.FileMode) (*os.File, *syscall.Flock_t, *UtmpError) {

	fi, err := os.OpenFile(UtmpFile, os.O_RDWR, os.ModeExclusive)
	if err != nil {
		return nil, nil, &UtmpError{err, nil}
	}
	//defer fi.Close()

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
		return nil, nil, &UtmpError{nil, err}
	}

	// Catch signals in case a terminating signal is sent before
	// we have a chance to unlock the file
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c,
			syscall.SIGTERM,
			syscall.SIGHUP,
			syscall.SIGQUIT,
			syscall.SIGINT)
		<-c

		lk.Type = syscall.F_UNLCK
		// We need to exit ASAPm so if the fcntl unlock fails exit anyway
		_ = syscall.FcntlFlock(fi.Fd(), syscall.F_SETLK, &lk)
		os.Exit(1)
	}()

	return fi, &lk, nil
}

// Unlocks the file and then closes it
func SafeClose(fi *os.File, lk *syscall.Flock_t) error {
	Unlock(fi, lk)
	err := fi.Close()
	if err != nil {
		return err
	}
	return nil
}

// Unlock open file
func Unlock(fi *os.File, lk *syscall.Flock_t) {
	lk.Type = syscall.F_ULOCK
	_ = syscall.FcntlFlock(fi.Fd(), syscall.F_SETLK, lk)
}

// Rewind to beginning of file
func SetUtEnd(fi *os.File) error {
	_, err := fi.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}
	return nil
}

// Searches forward from point in file and finds the correct entry based on id
// Returns -1 if no appropriate entry is found
func GetUtid(fi *os.File, id *Utmp) (int64, *Utmp) {
	var i int64

	// I do this because
	// https://github.com/lattera/glibc/blob/master/login/getutid_r.c#L43
	// I'm sure '<' and '>' would work just fine in Go, however.
	if id.Type != RunLevel &&
		id.Type != BootTime &&
		id.Type != NewTime &&
		id.Type != OldTime &&
		id.Type != InitProcess &&
		id.Type != LoginProcess &&
		id.Type != UserProcess &&
		id.Type != DeadProcess {
		return -1, nil
	}

	if id.Type == InitProcess ||
		id.Type == LoginProcess ||
		id.Type == UserProcess ||
		id.Type == DeadProcess {
		goto Process
	}

	if id.Type == RunLevel ||
		id.Type == BootTime ||
		id.Type == NewTime ||
		id.Type == OldTime {
		goto Time
	}

Process:
	i = int64(0)
	for {
		u := new(Utmp)

		err := binary.Read(fi, binary.LittleEndian, u)
		if err != nil && err != io.EOF {
			break
		}
		if err == io.EOF {
			break
		}

		if u.Id == id.Id {
			return (i + 1) * int64(unsafe.Sizeof(u)), u
		}
		i++
	}

Time:
	i = int64(0)
	for {
		u := new(Utmp)

		err := binary.Read(fi, binary.LittleEndian, u)
		if err != nil && err != io.EOF {
			break
		}
		if err == io.EOF {
			break
		}

		if u.Type == id.Type {
			return (i + 1) * int64(unsafe.Sizeof(u)), u
		}
		i++
	}

	return -1, nil
}

// Write to a wtmp file.
// On error returns a pointer to a UtmpError struct, else nil
func WriteWtmp(fi *os.File, lk *syscall.Flock_t, user, id string, pid int32, utype int16, line string) *UtmpError {

	u := new(Utmp)
	u.Time.GetTimeOfDay()
	u.Pid = pid
	u.Type = utype
	_ = copy(u.User[:], []byte(user))
	_ = copy(u.Id[:], []byte(id))
	_ = copy(u.Line[:], []byte(line))

	name := new(syscall.Utsname)
	if syscall.Uname(*&name) == nil {
		// gen.Int8toByte65 in gen_helper_funcs.go
		_ = copy(u.Host[:], gen.Int8toByte(name.Release[:]))
	}

	err := UpdWtmp(fi, lk, u)
	if err != nil {
		return &UtmpError{err.WriteErr, nil}
	}

	e := SafeClose(fi, lk)
	if e != nil {
		return &UtmpError{nil, e}
	}

	return nil
}

// Write to a utmp file.
// On error returns a pointer to a UtmpError struct, else nil
func WriteUtmp(fi *os.File, lk *syscall.Flock_t, user, id string, pid int32, utype int16, line, oldline string) *UtmpError {

	u := new(Utmp)
	u.Time.GetTimeOfDay()
	u.Pid = pid
	u.Type = utype
	_ = copy(u.User[:], []byte(user))
	_ = copy(u.Id[:], []byte(id))
	_ = copy(u.Line[:], []byte(line))

	name := new(syscall.Utsname)
	if syscall.Uname(*&name) == nil {
		// gen.Int8toByte65 in gen_helper_funcs.go
		_ = copy(u.Host[:], gen.Int8toByte(name.Release[:]))
	}

	if utype == DeadProcess {
		if r, st := GetUtid(fi, u); r > -1 {
			_ = copy(u.Line[:], st.Line[:])
			if oldline != "" {
				_ = copy([]byte(oldline), st.Line[:])
			}
		}
	}

	err := SetUtEnd(fi)
	if err != nil {
		return &UtmpError{err, nil}
	}

	if err := u.PutUtLine(fi, lk); err != nil {
		return &UtmpError{err.WriteErr, nil}
	}

	err = SafeClose(fi, lk)
	if err != nil {
		return &UtmpError{nil, err}
	}

	return nil
}

// Writes to UtmpFile at fi's current position, else append
func (u *Utmp) PutUtLine(fi *os.File, lk *syscall.Flock_t) *UtmpError {
	su := unsafe.Sizeof(u)

	// Save current position
	cur, _ := GetUtid(fi, u)

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

	if cur == -1 {
		cur = sz
	}
	_, err = fi.Seek(cur, os.SEEK_SET)
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

func WriteUtmpWtmp(fi *os.File, lk *syscall.Flock_t, user, id string, pid int32, utype int16, line string) bool {
	var oldline string

	if user == "" {
		return false
	}

	WriteUtmp(fi, lk, user, id, pid, utype, line, oldline)
	if line == "" && line[0] == 0 {
		line = oldline
	}
	WriteWtmp(fi, lk, user, id, pid, utype, line)

	return true
}
