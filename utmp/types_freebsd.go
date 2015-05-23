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

// #include <utmpx.h>
// #include "utxdb.h"
import "C"

const TypeNotDefined = false

// Values for Utmp.Type field
const (
	Empty        = C.EMPTY         // No valid user accounting information.
	BootTime     = C.BOOT_TIME     // Time of system boot.
	OldTime      = C.OLD_TIME      // Time when system clock changed.
	NewTime      = C.NEW_TIME      // Time after system clock changed.
	UserProcess  = C.USER_PROCESS  // A process.
	InitProcess  = C.INIT_PROCESS  // A process spawned by the init process.
	LoginProcess = C.LOGIN_PROCESS // Identifies the session leader of a logged-in user.
	DeadProcess  = C.DEAD_PROCESS  // A session leader who has exited.
	ShutdownTime = C.SHUTDOWN_TIME // Time of system shutdown.
)

// utmp, wtmp, btmp, and lastlog file names
const (
	// Usually /var/run/utx.active
	UtxActive = C._PATH_UTX_ACTIVE

	// Usually /var/log/utx.lastlogin
	UtxLastLog = C._PATH_UTX_LASTLOGIN

	// Usually /var/log/utx.log
	UtxLog = C._PATH_UTX_LOG
)

type Utmacro int

const (
	User Utmacro = iota
	Line
	Host
)

// DB status options
const (
	UtxDBActive    = C.UTXDB_ACTIVE
	UtxDBLastLogin = C.UTXDB_LASTLOGIN
	UtxDBLog       = C.UTXDB_LOG
)

// Structure describing the status of a terminated process.
type exit struct {
	Termination int16
	Exit        int16
}

// Not using syscall because int64s mess up our binary reads
type TimeVal struct {
	Sec  int32
	Usec int32
}

// The structure describing an entry in the database of prvious logins
type LastLog struct {
	Time int32
	Line [32]byte
	Host [128]byte
}

// The structure describing an entry in the user accounting database
type Utmpx struct {
	Type   int16     // Type of entry
	Time   TimeVal   // Time entry was made
	Id     [8]byte   // Terminal name suffix or inittab(5) ID
	Pid    int32     // Process ID
	User   [32]byte  // User login name
	Line   [16]byte  // Device name
	Host   [128]byte // Remote hostname
	Unused [64]byte  // Reserved for future use
}

type Futx struct {
	Type uint8     // Type of entry
	_    uint8     // padding
	Time uint64    // Time entry was made
	Id   [8]byte   // Terminal name suffix or intittab(5) ID
	Pid  uint32    // Process ID
	User [32]byte  // User login name
	Line [16]byte  // Device name
	Host [128]byte // Remote hostname
}

// Exit    exit     // Exit status of a process marked as DeadProcess; not used by Linux init(1)
// Session int32    // Session ID (getsid(2)), used for windowing
// Time    TimeVal  // Time entry was made
// Addr    [4]int32 // Internet address of remote host; IPv4 address uses just Addr[0]
// Unused [20]byte  // Reserved for future use
