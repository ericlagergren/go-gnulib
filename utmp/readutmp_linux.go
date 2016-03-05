// Copyright (c) 2015 Eric Lagergren
// Use of this source code is governed by the LGPL 2.1 or later.

// This file implements GNU's readutmp.c written in Go

package utmp

import (
	"syscall"

	"github.com/EricLagergren/go-gnulib/util"
)

// IsDesirable determines whether the Utmp entry is desired by the user who
// asked for the specified options.
func (u *Utmp) IsDesirable(opts int) bool {
	userProc := u.IsUserProcess()
	if (opts&ReadUserProcess != 0) && !userProc {
		return false
	}
	return !((opts&CheckPIDs != 0) &&
		userProc && 0 < u.Pid &&
		(syscall.Kill(int(u.Pid), 0) == syscall.ESRCH))
}

// TypeEquals is the same as this C macro:
//
// UT_TYPE_EQUALS(V) ((V)->type)
func (u *Utmp) TypeEquals(v int16) bool {
	return u.Type == v
}

// IsUserProcess is the same as this C macro:
//
// # define IS_USER_PROCESS(U)                                    \
//   (UT_USER (U)[0]                                              \
//    && (UT_TYPE_USER_PROCESS (U)                                \
//        || (UT_TYPE_NOT_DEFINED && UT_TIME_MEMBER (U) != 0)))
func (u *Utmp) IsUserProcess() bool {
	return u.User[0] != 0 &&
		(u.TypeEquals(UserProcess) || (TypeNotDefined && u.Tv.Sec != 0))
}

// ExtractTrimmedName returns the stringified version of a username.
// It trims after first null byte
func (u *Utmp) ExtractTrimmedName() string {
	return string(u.User[:util.Clen(u.User[:])])
}

// ReadUtmp reads the Utmp file indicated by name.
//
// Returns an error if any reads fail without EOF, else nil
func ReadUtmp(name string, opts int) ([]*Utmp, error) {

	file, err := Open(name, Reading)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var (
		us []*Utmp
		u  *Utmp
	)

	for {
		u = GetUtEnt(file)
		if u == nil {
			break
		}
		if u.IsDesirable(opts) {
			us = append(us, u)
		}
	}
	return us, err
}
