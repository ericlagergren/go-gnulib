/*
	POSIX getutent(3) written in Go

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
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"unsafe"
)

// Rewind to beginning of file
func SetUtEnt(file *os.File) error {
	_, err := file.Seek(0, os.SEEK_SET)
	return err
}

// Close file
func EndUtEnt(file *os.File) error {
	return file.Close()
}

func (u *Utmp) GetUtEnt(file *os.File) (error, *Utmp) {
	un := new(Utmp)
	err := binary.Read(file, binary.LittleEndian, un)

	return err, un
}

// Searches forward from point in file and finds the correct entry based on id
// Returns -1 if no appropriate entry is found
func (u *Utmp) GetUtid(file *os.File) (int64, *Utmp) {

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

		return -1, nil
	}

	const size = int(unsafe.Sizeof(*u))
	offset := 0

	if u.Type == RunLevel ||
		u.Type == BootTime ||
		u.Type == NewTime ||
		u.Type == OldTime {

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
				break
			}
			offset += size
		}

	} else if u.Type == InitProcess ||
		u.Type == LoginProcess ||
		u.Type == UserProcess ||
		u.Type == DeadProcess {

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
				break
			}
			offset += size
		}
	}

	return -1, nil
}

func (u *Utmp) GetUtLine(file *os.File) (*Utmp, error) {
	for {
		nu := new(Utmp)

		err := binary.Read(file, binary.LittleEndian, nu)
		if err != nil && err != io.EOF {
			break
		}
		if err == io.EOF {
			break
		}

		if nu.Type == LoginProcess || nu.Type == UserProcess {
			if bytes.Equal(nu.Line[:], u.Line[:]) {
				return nu, nil
			}
		}
	}

	return nil, nil
}
