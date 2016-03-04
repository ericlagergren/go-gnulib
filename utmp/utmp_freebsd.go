// Copyright (c) 2015 Eric Lagergren
// Use of this source code is governed by the LGPL 2.1 or later.

// This file contains GNU's utmp.c.

package utmp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"time"
	"unsafe"

	"github.com/EricLagergren/go-gnulib/endian"
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

func (u *Utmpx) UToFString(f *Futx, typ Utmacro) {
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
func (u *Utmpx) UToFID(f *Futx) {
	copy(f.Id[:], u.Id[:])
}

// #define	UTOF_ID(ut, fu) do { \
// 	memcpy((fu)->fu_id, (ut)->ut_id,				\
// 	    MIN(sizeof (fu)->fu_id, sizeof (ut)->ut_id));		\
// } while (0)
func (u *Utmpx) UToFPID(f *Futx) {
	f.Pid = uint32(u.Pid)
}

// #define	UTOF_TYPE(ut, fu) do { \
// 	(fu)->fu_type = (ut)->ut_type;					\
// } while (0)
func (u *Utmpx) UToFType(f *Futx) {
	f.Type = uint8(u.Type)
}

// #define	UTOF_TV(fu) do { \
// 	struct timeval tv;						\
// 	gettimeofday(&tv, NULL);					\
// 	(fu)->fu_tv = htobe64((uint64_t)tv.tv_sec * 1000000 +		\
// 	    (uint64_t)tv.tv_usec);					\
// } while (0)
func (u *Utmpx) UToFTV(f *Futx) {
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
		u.UToFID(f)
		u.UToFString(f, User)
		u.UToFString(f, Line)
		u.UToFString(f, Host)
		u.UToFPID(f)
	case InitProcess:
		u.UToFID(f)
		u.UToFPID(f)
	case LoginProcess:
		u.UToFID(f)
		u.UToFString(f, User)
		u.UToFString(f, Line)
		u.UToFPID(f)
	case DeadProcess:
		u.UToFID(f)
		u.UToFPID(f)
	default:
		f.Type = Empty
		return
	}

	u.UToFType(f)
	u.UToFID(f)
}

func (f *Futx) FToUString(u *Utmpx, typ Utmacro) {
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
func (f *Futx) FToUID(u *Utmpx) {
	copy(u.Id[:], f.Id[:])
}

// #define	UTOF_ID(ut, fu) do { \
// 	memcpy((fu)->fu_id, (ut)->ut_id,				\
// 	    MIN(sizeof (fu)->fu_id, sizeof (ut)->ut_id));		\
// } while (0)
func (f *Futx) FToUPID(u *Utmpx) {
	u.Pid = int32(u.Pid)
}

// #define	UTOF_TYPE(ut, fu) do { \
// 	(fu)->fu_type = (ut)->ut_type;					\
// } while (0)
func (f *Futx) FToUType(u *Utmpx) {
	u.Type = int16(f.Type)
}

// #define	UTOF_TV(fu) do { \
// 	struct timeval tv;						\
// 	gettimeofday(&tv, NULL);					\
// 	(fu)->fu_tv = htobe64((uint64_t)tv.tv_sec * 1000000 +		\
// 	    (uint64_t)tv.tv_usec);					\
// } while (0)
func (f *Futx) FToUTV(u *Utmpx) {
	var t uint64
	t = endian.Be64toh(f.Time)
	u.Time.Sec = int32(t / 1000000)
	u.Time.Usec = int32(t % 1000000)
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
		f.FToUID(u)
		f.FToUString(u, User)
		f.FToUString(u, Line)
		f.FToUString(u, Host)
		f.FToUPID(u)
	case InitProcess:
		f.FToUID(u)
		f.FToUPID(u)
	case LoginProcess:
		f.FToUID(u)
		f.FToUString(u, User)
		f.FToUString(u, Line)
		f.FToUPID(u)
	case DeadProcess:
		f.FToUID(u)
		f.FToUPID(u)
	default:
		u.Type = Empty
		return u
	}

	f.FToUType(u)
	f.FToUID(u)
	return u
}

func (f *Futx) UtxActiveAdd() error {
	var (
		e, err  error
		partial = int64(-1)
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
				_, e = file.Seek(-int64(size), os.SEEK_CUR)
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
				partial -= int64(size)
			}
		}
	}

	// Didn't find a match, so use the partial match. If no
	// partial was found, append the new record.
	if partial != -1 {
		_, err = file.Seek(partial, os.SEEK_SET)
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

			if _, err := file.Seek(-int64(size), os.SEEK_CUR); err != nil {
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

		_, e = file.Seek(-int64(size), os.SEEK_CUR)
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
		return
	}
	defer SafeClose(file, lk)

	f := new(Futx)
	if stat, err := file.Stat(); err != nil &&
		stat.Size()%int64(unsafe.Sizeof(f)) != 0 {

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

	// Trim the entry's trailing null bytes.
	for l = len(fu); l > 0 && fu[l-1] == 0; l-- {
		// Empty
	}

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
