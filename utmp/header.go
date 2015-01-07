/*
	UTMP(5) header file

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

package utmp

// Values for Utmp.Type field
const (
	Empty        = 0 // Record does not contain valid info (formerly known as UT_UNKNOWN on Linux)s
	RunLevel     = 1 // Change in system run-level (see init(8))s
	BootTime     = 2 // Time of system boot (in timeVal)s
	NewTime      = 3 // Time after system clock change (in timeVal)s
	OldTime      = 4 // Time before system clock change (in timeVal)s
	InitProcess  = 5 // Process spawned by init(8)s
	LoginProcess = 6 // Session leader process for user logins
	UserProcess  = 7 // Normal processs
	DeadProcess  = 8 // Terminated processs
	Accounting   = 9 // Not implemented

	LineSize = 32
	NameSize = 32
	HostSize = 256
)

// utmp, wtmp, btmp, and lastlog file names
const (
	UtmpFile    = "/var/run/utmp"
	WtmpFile    = "/var/log/wtmp"
	BtmpFile    = "/var/log/btmp"
	LastLogFile = "/var/log/LastLog"
)

// Opts for ReadUtmp()
const (
	CheckPIDs       = 1
	ReadUserProcess = 2
)

// These are similar to xalloc(1)
type UtmpBuffer map[uint64]*Utmp
type LastLogBuffer map[int64]*LastLog

type timeVal struct {
	Sec  int32
	Usec int32
}

type exit struct {
	Termination int16
	Exit        int16
}

type LastLog struct {
	Time int32
	Line [LineSize]byte
	Host [HostSize]byte
}

type Utmp struct {
	Type    int16          // Type of record
	_       int16          // padding because Go doesn't 4-byte align
	Pid     int32          // PID of login process
	Line    [LineSize]byte // Device name of tty - "/dev/"
	Id      [4]byte        // Terminal name suffix or inittab(5) ID
	User    [NameSize]byte // Username
	Host    [HostSize]byte // Hostname for remote login or kernel version for run-level messages
	Exit    exit           // Exit status of a process marked as DeadProcess; not used by Linux init(1)
	Session int32          // Session ID (getsid(2)), used for windowing
	Time    timeVal        // Time entry was made
	Addr    [4]int32       // Internet address of remote host; IPv4 address uses just Addr[0]
	Unused  [20]byte       // Reserved for future use
}
