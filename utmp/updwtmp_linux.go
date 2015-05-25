/*
	GNU's UPDWTMP(3) written in Go

	Copyright (C) 2015 Eric Lagergren

	This is free software; you can redistribute it and/or
	modify it under the terms of the GNU Lesser General Public
	License as published by the Free Software Foundation; either
	version 2.1 of the License, or (at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Lesser General Public License for more details.

	You should have received a copy of the GNU Lesser General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

/* Written by Eric Lagergren <ericscottlagergren@gmail.com> */

package utmp

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// Appends a Wtmp entry to the wtmp file
func (u *Utmp) UpdWtmp(path string) error {
	const utmpSize = unsafe.Sizeof(*u)
	var fileSize int64

	file, lk, err := SafeOpen(path)
	if err != nil {
		goto done
	}

	fileSize, err = file.Seek(0, os.SEEK_END)
	if err != nil {
		// Cannot safely get file size in order to write
		goto done
	}

	// If we can't safely write, undo our changes and exit
	if fileSize%int64(utmpSize) != 0 {
		fileSize -= int64(utmpSize)

		terr := syscall.Ftruncate(int(file.Fd()), fileSize)

		if terr != nil {
			err = fmt.Errorf("database is an invalid size, truncate failed: %s", terr)
		} else {
			err = fmt.Errorf("database is an invalid size, rewound to %d", fileSize)
		}

		goto done
	}

	err = binary.Write(file, binary.LittleEndian, &u)

done:
	if file != nil {
		SafeClose(file, lk)
	}
	return err
}

// Constructs a struct using LINE, USER, HOST, current time,
// and current PID. Calls UdpWtmp() to append entry.
func LogWtmp(file, line, user, host string) error {
	u := new(Utmp)
	u.Time.GetTimeOfDay()
	u.Pid = int32(os.Getpid())
	_ = copy(u.Host[:], []byte(host))
	_ = copy(u.User[:], []byte(user))
	_ = copy(u.Line[:], []byte(line))

	return u.UpdWtmp(file)
}
