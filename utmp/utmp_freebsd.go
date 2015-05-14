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

	"github.com/EricLager/go-gnulib/math"
	"github.com/EricLagerg/go-gnulib/general"
)

func htobe64(x uint64) uint64 {
	return math.Bswap64(x)
}

// Same as syscall.Gettimeofday, except this uses int32 due to alignment
// issues in the Utmp structs.
func (t *TimeVal) GetTimeOfDay() {
	now := time.Now().Unix()
	t.Usec = int32(now / 1000)
	t.Sec = int32(now)
}

func (u *Utmpx) UtOfString(f *Futx, typ Utmacro) {
	switch typ {
	case User:
		copy(f.User[:], u.User[:])
	case Host:
		copy(f.Host[:], u.Host[:])
	case Line:
		copy(f.Line[:], u.Line[:])
	default:
		panic("invalid type")
	}
}

// #define	UTOF_STRING(ut, fu, field) do { \
// 	strncpy((fu)->fu_ ## field, (ut)->ut_ ## field,		\
// 	    MIN(sizeof (fu)->fu_ ## field, sizeof (ut)->ut_ ## field));	\
// } while (0)
func (u *Utmpx) UtOfId(f *Futx) {
	copy(f.Id[:], u.Id[:])
}

// #define	UTOF_ID(ut, fu) do { \
// 	memcpy((fu)->fu_id, (ut)->ut_id,				\
// 	    MIN(sizeof (fu)->fu_id, sizeof (ut)->ut_id));		\
// } while (0)
func (u *Utmpx) UtOfPid(f *Futx) {
	copy(f.Pid[:], u.Pid[:])
}

// #define	UTOF_TYPE(ut, fu) do { \
// 	(fu)->fu_type = (ut)->ut_type;					\
// } while (0)
func (u *Utmpx) UtOfType(f *Futx) {
	copy(f.Type[:], u.Type[:])
}

// #define	UTOF_TV(fu) do { \
// 	struct timeval tv;						\
// 	gettimeofday(&tv, NULL);					\
// 	(fu)->fu_tv = htobe64((uint64_t)tv.tv_sec * 1000000 +		\
// 	    (uint64_t)tv.tv_usec);					\
// } while (0)
func (u *Utmpx) UtOfTv(f *Futx) {
	tv := new(TimeVal)
	tv.GetTimeOfDay()

}

func (u *Utmpx) UtxToFutx(f *Futx) {
	switch u.Type {
	case BootTime:
		fallthrough
	case OldTime:
		fallthrough
	case NewTime:
		fallthrough
	case ShutdownTime:
		break
	case UserProcess:

	}
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
func (u *Utmpx) PutUtLine(file *os.File) error {
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
