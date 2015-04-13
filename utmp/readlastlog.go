/*
	Read a lastlog file into a buffer

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
	"io"
	"os"
)

// Reads entries from a lastlog file into the LastLogBuffer specified by `ls`
// Returns an error if any reads fail without EOF
func ReadLastLog(entries *uint64, ls *[]LastLog) error {
	var e error

	fi, err := os.OpenFile(LastLogFile, os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		return err
	}
	defer fi.Close()

	i := uint64(0)
	for {
		var l LastLog

		err = binary.Read(fi, binary.LittleEndian, &l)
		if err != nil && err != io.EOF {
			e = err
		}
		if err == io.EOF {
			break
		}

		if len(*ls) <= int(i) {
			*ls = append(*ls, l)
		} else {
			(*ls)[i] = l
		}
		i++

		if i == *entries && i > 0 {
			break
		}
	}
	*entries = i

	return e
}
