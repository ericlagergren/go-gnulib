/*
	GNU's utmp.c written in Go

	Copyright (C) 2015 Eric Lagergren

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
	"syscall"
	"time"
	"unsafe"

	"github.com/EricLagerg/go-gnulib/general"
)

// Similar to glibc's Gettimeofday()
func (t *timeVal) GetTimeOfDay() {
	now := time.Now().Unix()
	t.Usec = int32(now / 1000)
	t.Sec = int32(now)
}

// A wrapper around os.OpenFile() that locks the file after opening
// Returns a pointer to the open fd, the lock struct, and an error/nil
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

// Unlocks the file and then closes it
func SafeClose(file *os.File, lk *syscall.Flock_t) error {
	Unlock(file, lk)
	err := file.Close()
	if err != nil {
		return err
	}
	return nil
}

// Unlock open file
func Unlock(file *os.File, lk *syscall.Flock_t) {
	lk.Type = syscall.F_ULOCK
	_ = syscall.FcntlFlock(file.Fd(), syscall.F_SETLK, lk)
}

// Rewind to beginning of file
func SetUtEnt(file *os.File) error {
	_, err := file.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}
	return nil
}

// Close file
func EndUtEnt(file *os.File) error {
	err := file.Close()
	if err != nil {
		return err
	}
	return nil
}

// Searches forward from point in file and finds the correct entry based on id
// Returns -1 if no appropriate entry is found
func (u *Utmp) GetUtid(file *os.File) (int64, *Utmp) {
	var i int64

	// I do this because
	// https://github.com/lattera/glibc/blob/master/login/getutid_r.c#L43
	// I'm sure '<' and '>' would work just fine in Go, however.
	if u.Type != RunLevel &&
		u.Type != BootTime &&
		u.Type != NewTime &&
		u.Type != OldTime &&
		u.Type != InitProcess &&
		u.Type != LoginProcess &&
		u.Type != UserProcess &&
		u.Type != DeadProcess {

		return -1, nil
	}

	if u.Type == RunLevel ||
		u.Type == BootTime ||
		u.Type == NewTime ||
		u.Type == OldTime {

		i := 0
		for {
			nu := new(Utmp)

			err := binary.Read(file, binary.LittleEndian, nu)
			if err != nil && err != io.EOF {
				break
			}
			if err == io.EOF {
				break
			}

			if u.Type == nu.Type {
				return (i + 1) * int(unsafe.Sizeof(*nu)), nu
			}
			i++
		}
		return -1, nil
	} else if u.Type == InitProcess ||
		u.Type == LoginProcess ||
		u.Type == UserProcess ||
		u.Type == DeadProcess {

		i := 0
		for {
			nu := new(Utmp)

			err := binary.Read(file, binary.LittleEndian, nu)
			if err != nil && err != io.EOF {
				break
			}
			if err == io.EOF {
				break
			}

			if u.Id == u.Id {
				return (i + 1) * int(unsafe.Sizeof(*nu)), nu
			}
			i++
		}
		return -1, nil
	}

	return -1, nil
}

// Write to a wtmp file.
// On error returns a pointer to a error struct, else nil
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

	err := UpdWtmp(file, u)
	if err != nil {
		return err
	}

	return nil
}

// Write to a utmp file.
// On error returns a pointer to a UtmpError struct, else nil
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

	if err := u.PutUtLine(file); err != nil {
		return err
	}

	return nil
}

// Writes to name at fi's current position, else append
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

	err = binary.Write(file, binary.LittleEndian, u)
	if err != nil {
		return err
	}

	return nil
}

func WriteUtmpWtmp(file *os.File, user, id string, pid int32, utype int16, line string) bool {
	var oldline string

	if user == "" {
		return false
	}

	WriteUtmp(file, user, id, pid, utype, line, oldline)
	if line == "" && line[0] == 0 {
		line = oldline
	}
	WriteWtmp(file, user, id, pid, utype, line)

	return true
}
