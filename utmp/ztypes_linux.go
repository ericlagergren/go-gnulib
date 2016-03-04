// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs types_linux.go

package utmp

import (
	"time"
	"unsafe"
)

const (
	TypeNotDefined = 0x1 == 0
	UserSize       = 0x4
)

const (
	CheckPIDs       = 0x1
	ReadUserProcess = 0x2
)

const BufSiz = 0x2000

type ExitStatus struct {
	X__e_termination int16
	X__e_exit        int16
}

type TimeVal struct {
	Sec  int32
	Usec int32
}

func (t *TimeVal) GetTimeOfDay() {
	now := time.Now().Unix()
	t.Sec = int32(now)
	t.Usec = int32(now / 1000)
}

const (
	Linesize = 0x20
	Namesize = 0x20
	Hostsize = 0x100
)

type Utmp struct {
	Type              int16
	Pad_cgo_0         [2]byte
	Pid               int32
	Line              [32]byte
	Id                [4]byte
	User              [32]byte
	Host              [256]byte
	Exit              ExitStatus
	Session           int32
	Tv                TimeVal
	Addr_v6           [4]int32
	X__glibc_reserved [20]byte
}

type LastLog struct {
	Time int32
	Line [32]byte
	Host [256]byte
}

const (
	Empty        = 0x0
	RunLevel     = 0x1
	BootTime     = 0x2
	NewTime      = 0x3
	OldTime      = 0x4
	InitProcess  = 0x5
	LoginProcess = 0x6
	UserProcess  = 0x7
	DeadProcess  = 0x8
	Accounting   = 0x9
	Unknown      = 0x0
)

const (
	UtmpxFile   = "/var/run/utmp"
	Wtmpxfile   = "/var/log/wtmp"
	LastLogFile = "/var/log/lastlog"
)

const utmpSize = unsafe.Sizeof(Utmp{})
