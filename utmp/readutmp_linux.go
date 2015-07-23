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
	"os"
	"syscall"

	"github.com/EricLagerg/go-gnulib/general"
)

// Determines whether the Utmp entry is desired by the user who asked for
// the specified options
func (u *Utmp) IsDesirable(opts int) bool {
	userProc := u.IsUserProcess()
	if (opts&ReadUserProcess != 0) && !userProc {
		return false
	}

	if (opts&CheckPIDs != 0) &&
		userProc &&
		0 < u.Pid &&
		(syscall.Kill(int(u.Pid), 0) == syscall.ESRCH) {

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
func ReadUtmp(fname string, entries *uint64, size int, opts int) ([]*Utmp, error) {
	var err error

	file, err := os.OpenFile(fname, os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var (
		us   = make([]*Utmp, 0, size)
		u    *Utmp
		read = uint64(0)
	)

	for {
		if u = GetUtEnt(file); u == nil {
			break
		}

		if u.IsDesirable(opts) {
			us = append(us, u)
			read++
		}
	}

	*entries = read
	return us, err
}
