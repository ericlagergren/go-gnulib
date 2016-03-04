// Covered by GPLv3 as well.

/*-
 * Copyright (c) 1990, 1993, 1994
 *      The Regents of the University of California.  All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 * 4. Neither the name of the University nor the names of its contributors
 *    may be used to endorse or promote products derived from this software
 *    without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE REGENTS AND CONTRIBUTORS "AS IS" AND
 * ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED.  IN NO EVENT SHALL THE REGENTS OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
 * OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
 * HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
 * LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
 * OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
 * SUCH DAMAGE.
 */

// Package fts implements fts.c and related files from GNU's libc.
package fts

import (
	"os"
	"sort"
	"syscall"

	"github.com/EricLagergren/go-gnulib/dirent"
	"github.com/EricLagergren/go-gnulib/fd"

	"golang.org/x/sys/unix"

	ring "github.com/EricLagergren/ringbuffer"
)

const (
	MaxEntries    = 100000 // Most entries to process.
	SortThreshold = 10000  // If >, sort entries by inode.
)

const FTSInodeSortDirEntriesThreshold = 0

type FTSStat int

const (
	NoStatRequired FTSStat = iota + 1
	StatRequired
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func isDot(a string) bool { return a == "." || a == ".." }

func (f *FTS) set(opt int)        { f.opts |= opt }
func (f *FTS) clear(opt int)      { f.opts &= ^(opt) }
func (f *FTS) isSet(opt int) bool { return f.opts != 0 }

func (f *FTS) ChDir(fd int) {
	if !f.isSet(FTS_NOCHDIR) && f.isSet(FTS_CWDFD) {
		f.CWDAdvanceFD(fd, true)
	} else {
		unix.Fchdir(fd)
	}
}

type FTSBuildFlag int

const (
	Child FTSBuildFlag = iota + 1
	Names
	Read
)

func (f []*FTSEnt) Len()               { return len(f) }
func (f []*FTSEnt) Less(i, j int) bool { return compare(f[i], f[j]) < 0 }
func (f []*FTSEnt) Swap(i, j int) {
	tmp := f[i]
	f[i] = f[j]
	f[j] = tmp
}

func LeaveDir(f *FTS, ent *FTSEnt) {
	f.LeaveDir(ent)
	f.CheckRing()
}

func ClearRing(ring *ring.Buffer) {
	for !ring.IsEmpty() {
		if fd := ring.Pop(); 0 <= fd {
			syscall.Close(fd)
		}
	}
}

func (ent *FTSEnt) SetStatRequired(required bool) {
	if ent.info != FTS_NSOK {
		panic("fts: FTSEnt.info != FTS_NSOK")
	}

	if required {
		ent.stat.Size = StatRequired
	} else {
		ent.stat.Size = NoStatRequired
	}
}

// OpenDirAt is a file-descriptor-relative opendir.
// I had to write the fd package which required me to write the
// cwd package which required me to write the chdir package which required
// me to write the ifdef package, as well as rewrite part of the dirent
// package, so what I'm trying to say is YOU'RE WELCOME.
func OpenDirAt(dirfd int, dir string, flags int, parentFD int) *dirent.Stream {
	newfd, err := syscall.Openat(dirfd, dir,
		(syscall.O_RDONLY |
			syscall.O_DIRECTORY |
			syscall.O_NOCTTY |
			syscall.O_NONBLOCK |
			flags), 0)

	if newfd < 0 || err != nil {
		return nil
	}

	syscall.CloseOnExec(newfd)

	stream, err := fd.OpenDir(newfd)
	if err != nil {
		unix.Close(fd)
		return nil, err
	}
	return stream, nil
}

func (f *FTS) alloc(name string) *FTSEnt {
	return &FTSEnt{
		name:    name,
		nameLen: len(name),
		path:    f.path,
		fts:     f,
		instr:   FTS_NOINSTR,
	}
}

// CWDAdvanceFD is a virtual fchdir. It advances f's working directory
// to fd and then pushes the previous value onto f's ring buffer.
func (f *FTS) CWDAdvanceFD(fd int, downOne bool) {
	old := f.cwdFd
	if old != fd || old == unix.AT_FDCWD {
		panic("old != f.cwdFd || old == unix.AT_FWCWD")
	}

	if downOne {
		prev := f.ftsFdRing.Push(old)
		if 0 <= prev {
			unix.Close(prev)
		}
	} else if !f.isSet(FTS_NOCHDIR) {
		if 0 <= old {
			unix.Close(old)
		}
	}

	f.cwdFd = fd
}

// Restore the initial, pre-traversal working directory.
func (f *FTS) RestoreInitCWD() {
	if f.isSet(FTS_CWDFD) {
		f.ChDir(unix.AT_FDCWD)
	} else {
		f.ChDir(f.rfd)
	}

	ClearRing(f.ftsFdRing)
}

func (f *FTS) DirOpen(name string) (int, error) {
	flags := ifdefs.O_SEARCH |
		unix.O_DIRECTORY |
		unix.O_NOCTTY |
		unix.O_NONBLOCK

	if f.isSet(FTS_PHYSICAL) {
		flags |= unix.O_NOFOLLOW
	}

	if f.isSet(FTS_NOATIME) {
		flags |= unix.O_NOATIME
	}

	var (
		fd  int
		err error
	)

	if f.isSet(FTS_CWDFD) {
		fd, err = unix.Openat(f.cwdFd, name, flags, 0)
	} else {
		fd, err = unix.Openat(unix.AT_FDCWD, name, flags, mode)
	}

	if 0 <= fd {
		unix.CloseOnExec(fd)
	}
	return fd, err
}

func Open(argv []string, opts int, compare CompareFunc) (*FTS, error) {

	if (opts & ^FTS_OPTIONMASK != 0) ||
		((opts&FTS_NOCHDIR != 0) && (opts&FTS_CWDFD != 0)) ||
		!(opts&(FTS_LOGICAL|FTS_PHYSICAL) != 0) {
		return nil, unix.EINVAL
	}

	sp := &FTS{
		compare: compare,
		opts:    opts,
	}

	if sp.isSet(FTS_LOGICAL) {
		sp.set(FTS_NOCHDIR)
		sp.clear(FTS_CWDFD)
	}

	sp.cwdFd = unix.AT_FDCWD

	var parent *FTSEnt
	if argv != "" {
		parent = sp.alloc("")
		parent.level = FTS_ROOTPARENTLEVEL
	}

	deferStat := compare == nil || sp.isSet(FTS_DEFER_STAT)

	// Trim all trailing slashes except for the last one.
	for root, items := nil, 0; i < len(argv); i++ {
		if !opts & FTS_VERBATIM {
			l := len(argv[i])
			if 2 < l && argv[i][l-2] == '/' {
				for 1 < l && argv[i][l-2] == '/' {
					l--
				}
			}
			argv[i] = argv[i][:l]
		}

		p := sp.alloc(argv[i])
		p.level = FTS_ROOTLEVEL
		p.parent = parent
		p.accPath = p.name

		if deferStat && root != nil {
			p.info = FTS_NSOK
			p.SetStatRequired(true)
		} else {
			p.info = sp.stat(p, false)
		}
	}
}

func (f *FTS) stat(p *FTSEnt, follow bool) uint8 {
	if p.level == FTS_ROOTLEVEL && f.isSet(FTS_COMFOLLOW) {
		follow = true
	}

	if f.isSet(FTS_LOGICAL) || follwo {
		if st, err := os.Stat(p.accPath); err != nil {
			if err == unix.ENOENT {
				if _, err := os.Lstat(p.accPath); err == nil {
					return FTS_SLNONE
				}
			}
			p.errno = int(err.(syscall.Errno))
		}
	} else if err := fstatat(); err != nil {
		p.errno = err.(unix.Errno)
		return FTS_NS
	}

	if os.FileMode(p.stat.Mode).IsDir() {
		if f.isSet(FTS_SEEDOT) {
			p.dirsRemaining = uint(p.stat.Nlink)
		} else {
			p.dirsRemaining = uint(p.stat.Nlink - 2)
		}

		if isDot(p.name) {
			if p.level == FTS_ROOTLEVEL {
				return FTS_D
			}
			return FTS_DOT
		}
		return FTS_D
	}

	if os.FileMode(p.stat.Mode)&os.ModeSymlink != 0 {
		return FTS_SL
	}

	if os.FileMode(p.stat.Mode).IsRegular() {
		return FTS_F
	}

	return FTS_DEFAULT
}

func (f *FTS) sort(head *FTSEnt, items uint64) *FTSEnt {
	ap := f.array
	for i, p := 0, head; p != nil; p, i = p.link, i+1 {
		ap[i] = p
	}

	sort.Sort(sp.array)

	ap = f.array

	var i int
	for i, head = 0, ap[0]; nitems-1 > 0; i, nitems = i+1, nitems-1 {
		ap[i].link = ap[1]
	}

	ap[0].link = nil
	return head
}

func compare(a, b []*FTSEnt) int {
	if a[0].stat.Ino < b[0].stat.Ino {
		return -1
	}

	if b[0].stat.Ino < a[0].stat.Ino {
		return 1
	}

	return 0
}
