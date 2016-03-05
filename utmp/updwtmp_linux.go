// Copyright (c) 2015 Eric Lagergren
// Use of this source code is governed by the LGPL 2.1 or later.

// This file contains GNU's updwtmp(3).

package utmp

import (
	"encoding/binary"
	"fmt"
	"os"
)

// UpdWtmp appends a Wtmp entry to the WTMP file.
func (u *Utmp) UpdWtmp(path string) error {
	file, err := Open(path, Writing)
	if err != nil {
		return err
	}
	defer file.Close()

	fileSize, err := file.Seek(0, os.SEEK_END)
	if err != nil {
		return err
	}

	if fileSize%int64(utmpSize) != 0 {
		fileSize -= int64(utmpSize)

		err := file.Truncate(fileSize)
		if err != nil {
			return fmt.Errorf("database is an invalid size, truncate failed: %v", err)
		}
		return fmt.Errorf("database is an invalid size, rewound to %d", fileSize)
	}

	err = binary.Write(file, binary.LittleEndian, &u)
	if err != nil {
		return file.Truncate(fileSize)
	}
	return nil
}

// LogWtmp constructs a struct using line, user, host, the current time,
// and current PID. Calls UdpWtmp() to append entry.
func LogWtmp(file, line, user, host string) error {
	var u Utmp
	u.Tv.GetTimeOfDay()
	u.Pid = pid
	_ = copy(u.Host[:], host)
	_ = copy(u.User[:], user)
	_ = copy(u.Line[:], line)
	return u.UpdWtmp(file)
}
