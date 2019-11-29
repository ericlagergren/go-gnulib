// Copyright (c) 2015 Eric Lagergren
// Use of this source code is governed by the LGPL 2.1 or later.

// This file contains GNU's utmp.c.

package utmp

import (
	"encoding/binary"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// WriteWtmp writes an event into the Wtmp file. An error is returned if the
// event cannot be appended to the Wtmp file.
func WriteWtmp(user, id string, pid int32, typ int16, line string) error {

	var u Utmp
	u.Tv.GetTimeOfDay()
	u.Pid = pid
	u.Type = typ
	_ = copy(u.User[:], []byte(user))
	_ = copy(u.Id[:], []byte(id))
	_ = copy(u.Line[:], []byte(line))

	var name unix.Utsname
	err := unix.Uname(&name)
	if err != nil {
		return err
	}
	_ = copy(u.Host[:], name.Release[:])
	return u.UpdWtmp(Wtmpxfile)
}

// WriteUtmp writes an event into the Utmp file. An error is returned if the
// event cannot be written to the Utmp file.
func WriteUtmp(user, id string, pid int32, typ int16, line string, oldline *string) error {

	var u Utmp
	u.Pid = pid
	u.Type = typ
	u.Tv.GetTimeOfDay()
	_ = copy(u.User[:], user)
	_ = copy(u.Id[:], id)
	_ = copy(u.Line[:], line)

	var name unix.Utsname
	if err := unix.Uname(&name); err == nil {
		_ = copy(u.Host[:], name.Release[:])
	}

	file, err := Open(UtmpxFile, Writing)
	if err != nil {
		return err
	}
	defer file.Close()

	if typ == DeadProcess {
		err := SetUtEnt(file)
		if err != nil {
			return err
		}
		if st, r := u.GetUtid(file); r > -1 {
			_ = copy(u.Line[:], st.Line[:])
			if oldline != nil {
				*oldline = string(st.Line[:])
			}
		}
	}

	err = SetUtEnt(file)
	if err != nil {
		return err
	}
	return u.PutUtLine(file)
}

// PutUtLine writes u to the file at the appropriate place in the database.
func (u *Utmp) PutUtLine(file *File) error {
	// Save current position
	_, cur := u.GetUtid(file)

	fileSize, err := file.Seek(0, os.SEEK_END)
	if err != nil {
		return err
	}

	// If we can't write safely undo our changes and exit
	if fileSize%int64(utmpSize) != 0 {
		fileSize -= int64(utmpSize)

		err := file.Truncate(fileSize)
		if err != nil {
			return fmt.Errorf("database is an invalid size, truncate failed: %v", err)
		}
		return fmt.Errorf("database is an invalid size, rewound to %d", fileSize)
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

// WriteUtmpWtmp writes to both Utmp and Wtmp files.
func WriteUtmpWtmp(file *File, user, id string, pid int32, typ int16, line string) {
	if user == "" {
		return
	}

	var oldline string
	WriteUtmp(user, id, pid, typ, line, &oldline)
	if line == "" && line[0] != 0 {
		line = oldline
	}
	WriteWtmp(user, id, pid, typ, line)
}
