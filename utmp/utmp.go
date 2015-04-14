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
	"fmt"
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
func Unlock(file *os.File, lk *syscall.Flock_t) error {
	if lk == nil || file == nil {
		return fmt.Errorf("file or lock are nil file: %s lk: %s", file, lk)
	}

	lk.Type = syscall.F_ULOCK
	return syscall.FcntlFlock(file.Fd(), syscall.F_SETLK, lk)
}

// Write an event into the Wtmp file. An error is returned if the event
// cannot be appended to the Wtmp file.
func WriteWtmp(user, id string, pid int32, utype int16, line string) error {

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

	return u.UpdWtmp(WtmpFile)
}

// Write an event into the Utmp file. An error is returned if the event
// cannot be written to the Utmp file.
func WriteUtmp(user, id string, pid int32, utype int16, line, oldline string) error {

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

	file, lk, err := SafeOpen(UtmpFile)
	if err != nil {
		goto done
	}

	if utype == DeadProcess {

		_ = SetUtEnt(file)
		if r, st := u.GetUtid(file); r > -1 {
			_ = copy(u.Line[:], st.Line[:])
			if oldline != "" {
				_ = copy([]byte(oldline), st.Line[:])
			}
		}
	}

	err = SetUtEnt(file)
	if err != nil {
		goto done
	}

	err = u.PutUtLine(file)

done:
	if file != nil {
		SafeClose(file, lk)
	}
	return err
}

// Writes to name at the appropriate place in the database
func (u *Utmp) PutUtLine(file *os.File) error {
	const utmpSize = unsafe.Sizeof(*u)

	// Save current position
	cur, _ := u.GetUtid(file)

	fileSize, err := file.Seek(0, os.SEEK_END)
	if err != nil {
		// Cannot safely get file size in order to write
		return err
	}

	// If we can't write safely undo our changes and exit
	if fileSize%int64(utmpSize) != 0 {
		fileSize -= int64(utmpSize)

		terr := syscall.Ftruncate(int(file.Fd()), fileSize)

		if terr != nil {
			err = fmt.Errorf("database is an invalid size, truncate failed: %s", terr)
		} else {
			err = fmt.Errorf("database is an invalid size, rewound to %d", fileSize)
		}

		return err
	}

	if cur == -1 {
		cur = fileSize
	}
	_, err = file.Seek(cur, os.SEEK_SET)
	if err != nil {
		return err
	}

	return binary.Write(file, binary.LittleEndian, u)

}

// Write to both Utmp and Wtmp files
func WriteUtmpWtmp(file *os.File, user, id string, pid int32, utype int16, line string) {
	var oldline string

	if user == "" {
		return
	}

	WriteUtmp(user, id, pid, utype, line, oldline)
	if line == "" && line[0] == 0 {
		line = oldline
	}
	WriteWtmp(user, id, pid, utype, line)

	return
}
