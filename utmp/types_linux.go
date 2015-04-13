/*
	UTMP(5) types (header) file. Covers most utmp/wtmp functions.

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

// #include <utmp.h>
// #include <lastlog.h>
import "C"

// Values for Utmp.Type field
const (
	Empty        = C.EMPTY         // Record does not contain valid info (formerly known as UT_UNKNOWN on Linux)s
	RunLevel     = C.RUN_LVL       // Change in system run-level (see init(8))s
	BootTime     = C.BOOT_TIME     // Time of system boot (in timeVal)s
	NewTime      = C.NEW_TIME      // Time after system clock change (in timeVal)s
	OldTime      = C.OLD_TIME      // Time before system clock change (in timeVal)s
	InitProcess  = C.INIT_PROCESS  // Process spawned by init(8)s
	LoginProcess = C.LOGIN_PROCESS // Session leader process for user logins
	UserProcess  = C.USER_PROCESS  // Normal processs
	DeadProcess  = C.DEAD_PROCESS  // Terminated processs
	Accounting   = C.ACCOUNTING    // Not implemented
	Unknown      = C.EMPTY         // Old Linux name for Empty

	LineSize = C.UT_LINESIZE
	NameSize = C.UT_NAMESIZE
	HostSize = C.UT_HOSTSIZE
)

// utmp, wtmp, btmp, and lastlog file names
const (
	UtmpFile     = C._PATH_UTMP
	UtmpFileName = UtmpFile
	WtmpFile     = C._PATH_WTMP
	WtmpFileName = WtmpFile

	BtmpFile        = "/var/log/btmp"
	BtmpFileName    = BtmpFile
	LastLogFile     = "/var/log/lastlog"
	LastLogFileName = LastLogFile
)

// Options for ReadUtmp
const (
	CheckPIDs       = 1
	ReadUserProcess = 2
)

// Structure describing the status of a terminated process.
type exit struct {
	Termination int16
	Exit        int16
}

// Not using syscall because int64s mess up our binary reads
type timeVal struct {
	Sec  int32
	Usec int32
}

// The structure describing an entry in the database of prvious logins
type LastLog struct {
	Time int32
	Line [LineSize]byte
	Host [HostSize]byte
}

// The structure describing an entry in the user accounting database
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
