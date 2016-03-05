// +build ignore

// go tool cgo -godefs types_linux.go
// And then convert in8 to byte with search and replace.

package utmp

// #include <stdio.h>
// #include <utmpx.h>
// #include <lastlog.h>
// #include "readutmp.h"
import "C"

import (
	"time"
	"unsafe"
)

// Misc.
const (
	TypeNotDefined = C.UT_TYPE_NOT_DEFINED == 0
	UserSize       = C.UT_USER_SIZE
)

// Options for ReadUtmp
const (
	CheckPIDs       = C.READ_UTMP_CHECK_PIDS
	ReadUserProcess = C.READ_UTMP_USER_PROCESS
)

// Expose stdio's BUFSIZ variable.
const BufSiz = C.BUFSIZ

type ExitStatus C.struct___exit_status

type TimeVal C.struct___0

// Same as syscall.Gettimeofday, except this uses int32 due to alignment
// issues in the Utmp structs.
func (t *TimeVal) GetTimeOfDay() {
	now := time.Now().Unix()
	t.Sec = int32(now)
	t.Usec = int32(now / 1000)
}

const (
	Linesize = C.UT_LINESIZE
	Namesize = C.UT_NAMESIZE
	Hostsize = C.UT_HOSTSIZE
)

type Utmp C.struct_utmpx

type LastLog C.struct_lastlog

const (
	Empty        = C.EMPTY
	RunLevel     = C.RUN_LVL
	BootTime     = C.BOOT_TIME
	NewTime      = C.NEW_TIME
	OldTime      = C.OLD_TIME
	InitProcess  = C.INIT_PROCESS
	LoginProcess = C.LOGIN_PROCESS
	UserProcess  = C.USER_PROCESS
	DeadProcess  = C.DEAD_PROCESS
	Accounting   = C.ACCOUNTING
	Unknown      = C.UT_UNKNOWN
)

const (
	UtmpxFile   = C._PATH_UTMP
	Wtmpxfile   = C._PATH_WTMP
	LastLogFile = C._PATH_LASTLOG
)

const utmpSize = unsafe.Sizeof(Utmp{})
