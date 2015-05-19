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
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"time"
	"unsafe"

	"github.com/EricLagerg/go-gnulib/endian"
)

// Order is the current endianness for a system
var Order binary.ByteOrder

func init() {
	if endian.ByteOrder == endian.LittleEndian {
		Order = binary.LittleEndian
	} else {
		Order = binary.BigEndian
	}
}

// GetTimeOfDay is the same as syscall.Gettimeofday, except this uses int32
// due to alignment issues in the Utmp structs.
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
	f.Time = endian.Htobe64(uint64(u.Time.Sec*1000000) + uint64(tv.Usec))
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
		u.UtOfId(f)
		u.UtOfString(f, User)
		u.UtOfString(f, Line)
		u.UtOfString(f, Host)
		u.UtOfPid(f)
	case InitProcess:
		u.UtOfId(f)
		u.UtOfPid(f)
	case LoginProcess:
		u.UtOfId(f)
		u.UtOfString(f, User)
		u.UtOfString(f, Line)
		u.UtOfPid(f)
	case DeadProcess:
		u.UtOfId(f)
		u.UtOfPid(f)
	default:
		f.Type = Empty
		return
	}

	u.UtOfType(f)
	u.UtOfId(f)
}

func (f *Futx) FtOuString(u *Utmpx, typ Utmacro) {
	switch typ {
	case User:
		copy(u.User[:], f.User[:])
	case Host:
		copy(u.Host[:], f.Host[:])
	case Line:
		copy(u.Line[:], f.Line[:])
	default:
		panic("invalid type")
	}
}

// #define	UTOF_STRING(ut, fu, field) do { \
// 	strncpy((fu)->fu_ ## field, (ut)->ut_ ## field,		\
// 	    MIN(sizeof (fu)->fu_ ## field, sizeof (ut)->ut_ ## field));	\
// } while (0)
func (f *Futx) UtOfId(u *Utmpx) {
	copy(u.Id[:], f.Id[:])
}

// #define	UTOF_ID(ut, fu) do { \
// 	memcpy((fu)->fu_id, (ut)->ut_id,				\
// 	    MIN(sizeof (fu)->fu_id, sizeof (ut)->ut_id));		\
// } while (0)
func (f *Futx) FtOuPid(u *Utmpx) {
	copy(u.Pid[:], f.Pid[:])
}

// #define	UTOF_TYPE(ut, fu) do { \
// 	(fu)->fu_type = (ut)->ut_type;					\
// } while (0)
func (f *Futx) FtOuType(u *Utmpx) {
	copy(u.Type[:], f.Type[:])
}

// #define	UTOF_TV(fu) do { \
// 	struct timeval tv;						\
// 	gettimeofday(&tv, NULL);					\
// 	(fu)->fu_tv = htobe64((uint64_t)tv.tv_sec * 1000000 +		\
// 	    (uint64_t)tv.tv_usec);					\
// } while (0)
func (f *Futx) FtOuTv(u *Utmpx) {
	var t uint64
	t = endian.Be64toh(f.Time)
	u.Time.Sec = t / 1000000
	u.Time.Usec = t % 1000000
}

func (f *Futx) FutxToUtx() *Utmpx {
	u := new(Utmpx)

	switch f.Type {
	case BootTime:
		fallthrough
	case OldTime:
		fallthrough
	case NewTime:
		fallthrough
	case ShutdownTime:
		break
	case UserProcess:
		f.UtOfId(u)
		f.UtOfString(u, User)
		f.UtOfString(u, Line)
		f.UtOfString(u, Host)
		f.UtOfPid(u)
	case InitProcess:
		f.UtOfId(u)
		f.UtOfPid(u)
	case LoginProcess:
		f.UtOfId(u)
		f.UtOfString(u, User)
		f.UtOfString(u, Line)
		f.UtOfPid(u)
	case DeadProcess:
		f.UtOfId(u)
		f.UtOfPid(u)
	default:
		u.Type = Empty
		return
	}

	f.UtOfType(u)
	f.UtOfId(u)
}

func (f *Futx) UtxActiveAdd() error {
	var (
		e       error
		partial = -1
		ret     = 0
	)

	file, lk, err := SafeOpen(UtxActive)
	if err != nil {
		return err
	}
	defer SafeClose(file, lk)

	for {
		var fe Futx
		const size = unsafe.Sizeof(fe)

		err = binary.Read(file, Order, &fe)
		if err != nil && err != io.EOF {
			e = err
		}
		if err == io.EOF {
			break
		}

		switch fe.Type {
		case BootTime:
			// Leave these intact
		case UserProcess, InitProcess, LoginProcess:
			fallthrough
		case DeadProcess:
			if bytes.Equal(f.Id[:], fe.Id[:]) {
				ret, e = file.Seek(-size, os.SEEK_CUR)
				goto exact
			}

			if fe.Type != DeadProcess {
				break
			}

			fallthrough
		default:
			if partial == -1 {
				partial, e = file.Seek(0, os.SEEK_CUR)
			}

			if partial != -1 {
				partial -= size
			}
		}
	}

	// Didn't find a match, so use the partial match. If no
	// partial was found, append the new record.
	if partial != -1 {
		ret = file.Seek(partial, os.SEEK_SET)
	}

exact:
	// FreeBSD checks ret and sets error to errno if this
	// condition is true. We already set e to the correct
	// error value, so we don't need to test this.
	// if ret == -1 { e = e }
	if err := binary.Write(file, Order, f); err != nil {
		e = err
	}
	// We can also skip the other conditions because we've
	// set our return value in all the other places an error
	// could pop up.
	return e
}

func (f *Futx) UtxActiveRemove() error {
	var e error

	file, lk, err := SafeOpen(UtxActive)
	if err != nil {
		return err
	}
	defer SafeClose(file, lk)

	for {
		var fe Futx
		const size = unsafe.Sizeof(fe)

		err = binary.Read(file, Order, &fe)
		if err != nil && err != io.EOF {
			e = err
		}
		if err == io.EOF {
			break
		}

		switch fe.Type {
		case UserProcess, InitProcess:
		case LoginProcess:
			if bytes.Equal(f.Id[:], fe.Id[:]) {
				continue
			}

			if n, err := file.Seek(-size, os.SEEK_CUR); err != nil {
				e = err
			} else if err := binary.Write(file, Order, f); err != nil {
				e = err
			}
		}
	}

	return e
}

func (f *Futx) UtxActiveInit() {
	file, lk, err := SafeOpen(UtxActive)
	if err != nil {
		return
	}
	defer SafeClose(file, lk)

	// Init with a single boot record
	_ = binary.Write(file, Order, f)
}

func UtxActivePurge() {
	os.Truncate(UtxActive, 0)
}

func (f *Futx) UtxLastLoginAdd() error {
	var (
		e  error
		fe Futx
	)

	file, lk, err := SafeOpen(UtxLastLog)
	if err != nil {
		return err
	}
	defer SafeClose(file, lk)

	for {
		const size = unsafe.Sizeof(fe)

		err = binary.Read(file, Order, &fe)
		if err != nil && err != io.EOF {
			e = err
		}
		if err == io.EOF {
			break
		}

		if !bytes.Equal(f.User[:], fe.User[:]) {
			continue
		}

		_, e = file.Seek(-size, os.SEEK_CUR)
		break
	}

	if e == nil {
		return binary.Write(file, Order, &fe)
	}

	return e
}

func UtxLastLoginUpgrade() {
	file, lk, err := SafeOpen(UtxLastLog)
	if err != nil {
		return err
	}
	defer SafeClose(file, lk)

	f := new(Futx)
	if stat, err := file.Stat(); err != nil &&
		stat.Size()%unsafe.Sizeof(f) != 0 {

		file.Truncate(0)
	}
}

func (f *Futx) UtxLogAdd() error {
	var (
		e error
		l int
	)

	// Create temporary buffer to hold f and write
	// f as a byte slice.
	buf := make([]byte, unsafe.Sizeof(*f))
	b := bytes.NewBuffer(buf)
	e = binary.Write(b, Order, f)

	fu := b.Bytes()

	for l = len(fl); l > 0 && fu[l-1]; l-- {
		// Empty
	}
	l = endian.Htobe16(l)

	file, lk, err := SafeOpen(UtxLog)
	if err != nil {
		return err
	}
	defer SafeClose(file, lk)

	if e == nil {
		return binary.Write(file, Order, fu)
	}

	return e
}

// Writes to name at the appropriate place in the database
// Make sure to check the return value, because a non-nil
// return means something went wrong and is the equivalent
// of returning NULL on FreeBSD.
func (u *Utmpx) PutUtLine(file *os.File) error {
	var e error

	f := new(Futx)
	u.UtxToFutx(f)

	switch f.Type {
	case BootTime:
		f.UtxActiveInit()
		UtxLastLoginUpgrade()
	case ShutdownTime:
		UtxActivePurge()
	case OldTime:
		fallthrough
	case NewTime:
	case UserProcess:
		e = f.UtxActiveAdd()
		e = f.UtxLastLoginAdd()
	case DeadProcess:
		if e = f.UtxActiveRemove(); e != nil {
			return e
		}
	default:
		return errors.New("EINVAL")
	}

	e = f.UtxLogAdd()
	return e
}
