package utmp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"unsafe"
)

var (
	UDB   int
	UFile *os.File
)

func SetUtxDB(db int, file string) error {
	var err error

	switch db {
	case UtxDBActive:
		if file == "" {
			file = UtxActive
		}
	case UtxDBLastLogin:
		if file == "" {
			file = UtxLastLog
		}
	case UtxDBLog:
		if file == "" {
			file = UtxLog
		}
	default:
		return errors.New("EINVAL")
	}

	if UFile != nil {
		_ = UFile.Close()
	}

	UFile, err = os.Open(file)
	if err != nil {
		return err
	}

	if db != UtxDBLog {
		var fu Futx
		// Is the file broken?
		if stat, err := UFile.Stat(); err != nil &&
			uintptr(stat.Size())%unsafe.Sizeof(fu) != 0 {

			_ = UFile.Close()
			UFile = nil

			return errors.New("EFTYPE")
		}
		// setvbuf
	}

	UDB = db

	return nil
}

func SetUtxEnt() {
	SetUtxDB(UtxDBActive, "")
}

func EndUtxEnd() {
	if UFile != nil {
		UFile.Close()
		UFile = nil
	}
}

func (f *Futx) GetFutxEnt() error {
	if UFile == nil {
		SetUtxEnt()
	}

	if UFile == nil {
		return errors.New("Could not set global UFile")
	}

	if UDB == UtxDBLog {
		var length uint16

	retry:
		if err := binary.Read(UFile, Order, &length); err != nil {
			return err
		}

		length = endian.Be16toh(length)
		if length == 0 {
			// Seek forward one byte and try again
			UFile.Seek(int64(length+1), os.SEEK_CUR)
			goto retry
		}

		if uintptr(length) > unsafe.Sizeof(*f) {
			// Hell if I know...
			if err := binary.Read(UFile, Order, f); err != nil {
				return err
			}

			UFile.Seek(int64(uintptr(length)-unsafe.Sizeof(*f)),
				os.SEEK_CUR)
		} else {
			// Reset f because it's a partial record
			f = new(Futx)
			if err := binary.Read(UFile, Order, f); err != nil {
				return err
			}
		}
	} else {
		if err := binary.Read(UFile, Order, f); err != nil {
			return err
		}
	}

	return nil
}

func GetUtxEnt() *Utmpx {
	var fu Futx

	if fu.GetFutxEnt() != nil {
		return nil
	}

	return fu.FutxToUtx()
}

func (u *Utmpx) GetUtxId() *Utmpx {
	var fu Futx

	for {
		if fu.GetFutxEnt() != nil {
			return nil
		}

		switch fu.Type {
		case UserProcess:
			fallthrough
		case InitProcess:
			fallthrough
		case LoginProcess:
			fallthrough
		case DeadProcess:
			switch u.Type {
			case UserProcess:
				fallthrough
			case InitProcess:
				fallthrough
			case LoginProcess:
				fallthrough
			case DeadProcess:
				if bytes.Equal(fu.Id[:], u.Id[:]) {
					goto found
				}
			}
		default:
			if int16(fu.Type) == u.Type {
				goto found
			}
		}
	}

found:
	return fu.FutxToUtx()
}

func (u *Utmpx) GetUtxLine() *Utmpx {
	var fu Futx

	for {
		if fu.GetFutxEnt() != nil {
			return nil
		}

		switch fu.Type {
		case UserProcess:
			fallthrough
		case LoginProcess:
			if bytes.Equal(fu.Line[:], u.Line[:]) {
				goto found
			}
		}
	}

found:
	return fu.FutxToUtx()
}

func (u *Utmpx) GetUtxUser(user string) {
	var fu Futx
	bu := []byte(user)

	for {
		if fu.GetFutxEnt() != nil {
			return
		}

		switch fu.Type {
		case UserProcess:
			if bytes.Equal(fu.User[:], bu) {
				goto found
			}
		}
	}

found:
	u = fu.FutxToUtx()
}
