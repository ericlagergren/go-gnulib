/*
	GNU's readutmp.c written in Go

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
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"syscall"
)

var (
	nul = []byte{0}
)

func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}

// Determines whether the Utmp entry is desired by the user who asked for
// the specified options
func (u *Utmp) isDesirable(opts int) bool {
	p := u.IsUserProcess()
	if opts == 0 {
		return true
	}
	if (opts&ReadUserProcess != 0) && !p {
		return false
	}
	if (opts&CheckPIDs != 0) && p && 0 < u.Pid && syscall.Kill(int(u.Pid), 0) != syscall.ESRCH {
		return false
	}
	return true
}

// Basically the same as this C macro:
//
//# define IS_USER_PROCESS(U)                                     \
//   (UT_USER (U)[0]                                              \
//    && (UT_TYPE_USER_PROCESS (U)                                \
//        || (UT_TYPE_NOT_DEFINED && UT_TIME_MEMBER (U) != 0)))
func (u *Utmp) IsUserProcess() bool {
	return u.User[0] != 0 && (u.Type == UserProcess || u.Time.Sec != 0)
}

// Return stringified version of a username.
// Trims after first null ([]byte{0}) byte
func (u *Utmp) ExtractTrimmedName() string {
	return string(u.User[:clen(u.User[:])])
}

// Reads entries from a *tmp file and returns a channel of *Utmps
// Returns an error if any reads fail without EOF, else nil
func ReadUtmp(fname string, entries uint64, us *UtmpBuffer, opts int) error {
	var e error

	fi, err := os.OpenFile(fname, os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		return err
	}
	defer fi.Close()

	if entries == 0 {
		// Max unsigned int, so basically never ending
		entries = uint64(^(uint(0) >> 1))
	}
	i := uint64(0)
	for {
		u := new(Utmp)

		err = binary.Read(fi, binary.LittleEndian, u)
		if err != nil && err != io.EOF {
			e = err
			break
		}
		if err == io.EOF {
			break
		}

		if u.isDesirable(opts) {
			(*us)[i] = u
			i++
		}
		if i == entries {
			break
		}
	}
	return e
}
