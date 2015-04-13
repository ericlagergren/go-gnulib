/*
	GNU's utmp.c written in Go

	Copyright (C) 2015 Eric Lagergren

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 2 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

/* Written by Eric Lagergren <ericscottlagergren@gmail.com> */

// Package utmp provides the Go-equivalent of the POSIX-defined UTMP
// routines.
package utmp

import (
	"encoding/binary"
	"os"
	"syscall"
	"time"
	"unsafe"

	"github.com/EricLagerg/go-gnulib/general"
)

// Same as syscall.Gettimeofday, except this uses int32 due to alignment
// issues in the Utmp structs.
func (t *timeVal) GetTimeOfDay() {
	now := time.Now().Unix()
	t.Usec = int32(now / 1000)
	t.Sec = int32(now)
}

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
func Unlock(file *os.File, lk *syscall.Flock_t) {
	lk.Type = syscall.F_ULOCK
	_ = syscall.FcntlFlock(file.Fd(), syscall.F_SETLK, lk)
}

// Write an event into the Wtmp file. An error is returned if the event
// cannot be appended to the Wtmp file.
func WriteWtmp(file *os.File, user, id string, pid int32, utype int16, line string) error {

	u := new(Utmp)
	u.Time.GetTimeOfDay()
	u.Pid = pid
	u.Type = utype
	_ = copy(u.User[:], []byte(user))
	_ = copy(u.Id[:], []byte(id))
	_ = copy(u.Line[:], []byte(line))

	name := new(syscall.Utsname)
	if err := syscall.Uname(name); err == nil {
		_ = copy(u.Host[:], general.Int8ToByte(name.Release[:]))
	}

	return u.UpdWtmp(file)
}

// Write an event into the Utmp file. An error is returned if the event
// cannot be written to the Utmp file.
func WriteUtmp(file *os.File, user, id string, pid int32, utype int16, line, oldline string) error {

	u := new(Utmp)
	u.Time.GetTimeOfDay()
	u.Pid = pid
	u.Type = utype
	_ = copy(u.User[:], []byte(user))
	_ = copy(u.Id[:], []byte(id))
	_ = copy(u.Line[:], []byte(line))

	name := new(syscall.Utsname)
	if err := syscall.Uname(name); err == nil {
		_ = copy(u.Host[:], general.Int8ToByte(name.Release[:]))
	}

	if utype == DeadProcess {
		if r, st := u.GetUtid(file); r > -1 {
			_ = copy(u.Line[:], st.Line[:])
			if oldline != "" {
				_ = copy([]byte(oldline), st.Line[:])
			}
		}
	}

	err := SetUtEnt(file)
	if err != nil {
		return err
	}

	return u.PutUtLine(file)
}

// Writes to name at
func (u *Utmp) PutUtLine(file *os.File) error {
	const su = unsafe.Sizeof(*u)

	// Save current position
	cur, _ := u.GetUtid(file)

	sz, err := file.Seek(0, os.SEEK_END)
	if err != nil {
		// Cannot safely get file size in order to write
		return err
	}

	// If we can't write safely rewind the file and exit
	if sz%int64(su) != 0 {
		sz -= int64(su)
		err = syscall.Ftruncate(int(file.Fd()), sz)
		if err != nil {
			return err
		}
	}

	if cur == -1 {
		cur = sz
	}
	_, err = file.Seek(cur, os.SEEK_SET)
	if err != nil {
		return err
	}

	return binary.Write(file, binary.LittleEndian, u)
}

// In glibc this is void
func WriteUtmpWtmp(file *os.File, user, id string, pid int32, utype int16, line string) {
	var oldline string

	if user == "" {
		return
	}

	WriteUtmp(file, user, id, pid, utype, line, oldline)
	if line == "" && line[0] == 0 {
		line = oldline
	}
	WriteWtmp(file, user, id, pid, utype, line)

	return
}
