/*
	Read a lastlog file into a buffer

	Copyright (C) 2014 Eric Lagergren

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

/* Written by Eric Lagergren <ericscottlagergren@gmail.com> */

package gnulib

import (
	"encoding/binary"
	"io"
	"os"
)

// Reads entries from a lastlog file into the LastLogBuffer specified by `ls`
// Returns an error if any reads fail without EOF
func ReadLastLog(entries *int64, ls *LastLogBuffer) error {
	fi, err := os.OpenFile(LastLogFile, os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		return err
	}
	defer fi.Close()

	i := int64(0)
	for {
		l := new(LastLog)

		err = binary.Read(fi, binary.LittleEndian, l)
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF {
			break
		}

		(*ls)[i] = l
		i++
		if &i == entries {
			break
		}
	}
	return nil
}
