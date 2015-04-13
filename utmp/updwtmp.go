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
	"os"
	"syscall"
	"unsafe"
)

// Appends structure U to the wtmp file
func (u *Utmp) UpdWtmp(fi *os.File) error {
	const su = unsafe.Sizeof(*u)

	sz, err := fi.Seek(0, os.SEEK_END)
	if err != nil {
		// Cannot safely get file size in order to write
		return err
	}

	// If we can't write safely reset the file and exit
	if sz%int64(su) != 0 {
		sz -= int64(su)
		err = syscall.Ftruncate(int(fi.Fd()), sz)
		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return binary.Write(fi, binary.LittleEndian, &u)
}

// Constructs a struct using LINE, USER, HOST, current time,
// and current PID. Calls UdpWtmp() to append entry.
func LogWtmp(file *os.File, line, user, host string) error {
	u := new(Utmp)
	u.Time.GetTimeOfDay()
	u.Pid = int32(os.Getpid())
	_ = copy(u.Host[:], []byte(host))
	_ = copy(u.User[:], []byte(user))
	_ = copy(u.Line[:], []byte(line))

	return u.UpdWtmp(file)
}
