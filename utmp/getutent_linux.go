// Copyright (c) 2015 Eric Lagergren
// Use of this source code is governed by the LGPL 2.1 or later.

// This file implements POSIX getutent(3)

package utmp

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

// SetUtEnt rewinds back to the beginning of the file.
func SetUtEnt(file *File) error {
	_, err := file.Seek(0, os.SEEK_SET)
	return err
}

// EndUtEnt closes the file.
func EndUtEnt(file *File) error {
	return file.Close()
}

// GetUtEnt retrieves a Utmp entry from the file.
func GetUtEnt(file *File) *Utmp {
	var u Utmp
	if err := binary.Read(file, binary.LittleEndian, &u); err != nil {
		return nil
	}
	return &u
}

// GetUtid searches forward from the current point in file and finds the
// correct entry based on id. It returns -1 if no appropriate entry is found.
func (u *Utmp) GetUtid(file *File) (*Utmp, int64) {

	// These constants aren't guarenteed to be within a certain range,
	// so we can't check with '<' and '>'
	if u.Type != RunLevel &&
		u.Type != BootTime &&
		u.Type != NewTime &&
		u.Type != OldTime &&
		u.Type != InitProcess &&
		u.Type != LoginProcess &&
		u.Type != UserProcess &&
		u.Type != DeadProcess {

		return nil, -1
	}

	var offset int64
	var nu Utmp

	switch u.Type {
	case RunLevel, BootTime, NewTime, OldTime:
		for {
			err := binary.Read(file, binary.LittleEndian, &nu)
			if err != nil {
				return nil, -1
			}
			if nu.Type == u.Type {
				break
			}
			offset += int64(utmpSize)
		}
	case InitProcess, LoginProcess, UserProcess, DeadProcess:
		for {
			err := binary.Read(file, binary.LittleEndian, &nu)
			if err != nil {
				return nil, -1
			}
			if nu.Id == u.Id {
				break
			}
			offset += int64(utmpSize)
		}
	}
	return nil, offset
}

// GetUtLine finds the next line in the file whose Type member is
// UserProcess or LoginProcess and whose Line member == u's Line
// member.
func (u *Utmp) GetUtLine(file *File) (*Utmp, error) {
	var nu Utmp
	for {

		err := binary.Read(file, binary.LittleEndian, nu)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if nu.Type == LoginProcess || nu.Type == UserProcess {
			if bytes.Equal(nu.Line[:], u.Line[:]) {
				return &nu, nil
			}
		}
	}
	return nil, nil
}
