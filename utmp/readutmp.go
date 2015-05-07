/*
	GNU's readutmp.c written in Go

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
	"syscall"

	"github.com/EricLagerg/go-gnulib/general"
)

// Determines whether the Utmp entry is desired by the user who asked for
// the specified options
func (u *Utmp) IsDesirable(opts int) bool {
	p := u.IsUserProcess()
	if (opts&ReadUserProcess != 0) && !p {
		return false
	}

	if (opts&CheckPIDs != 0) &&
		p &&
		0 < u.Pid &&
		(syscall.Kill(int(u.Pid), 0) != syscall.ESRCH) {

		return false
	}

	return true
}

// UT_TYPE_EQUALS(V) ((V)->type)
func (u *Utmp) TypeEquals(v int16) bool {
	return u.Type == v
}

// Basically the same as this C macro:
// # define IS_USER_PROCESS(U)                                    \
//   (UT_USER (U)[0]                                              \
//    && (UT_TYPE_USER_PROCESS (U)                                \
//        || (UT_TYPE_NOT_DEFINED && UT_TIME_MEMBER (U) != 0)))
func (u *Utmp) IsUserProcess() bool {
	return u.User[0] != 0 &&
		(u.TypeEquals(UserProcess) || (TypeNotDefined && u.Time.Sec != 0))
}

// Return stringified version of a username.
// Trims after first null byte
func (u *Utmp) ExtractTrimmedName() string {
	return string(u.User[:general.Clen(u.User[:])])
}

// Reads entries from a Utmp file into a UtmpBuffer, us
// Returns an error if any reads fail without EOF, else nil
//
// This differs from GNU's because it assigns to the slice US even if
// an error is found.
//
// ReadUtmp asks for a pointer to a slice because if US is smaller than
// entries, we change the pointer and grow the slice as needed. This is
// generally the same speed as maps for smaller number of entries; for
// entries over 10,000 or so maps' speed plummets.
func ReadUtmp(fname string, entries *uint64, us *[]Utmp, opts int) error {
	var e error

	fi, err := os.OpenFile(fname, os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		return err
	}
	defer fi.Close()

	i := uint64(0)
	for {
		var u Utmp

		err = binary.Read(fi, binary.LittleEndian, &u)
		if err != nil && err != io.EOF {
			e = err
		}
		if err == io.EOF {
			break
		}

		if u.IsDesirable(opts) {
			// Grow if needed
			if len(*us) <= int(i) {
				*us = append(*us, u)
			} else {
				(*us)[i] = u
			}

			i++
		}

		if i == *entries && i > 0 {
			break
		}
	}
	*entries = i

	return e
}
