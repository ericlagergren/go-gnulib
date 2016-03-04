/*
 * Copyright (c) 1989, 1993
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
 *
 *      @(#)fts.h       8.3 (Berkeley) 8/14/94
 */

package fts

import (
	"syscall"

	"github.com/EricLagergren/go-gnulib/dirent"
	rb "github.com/EricLagergren/ringbuffer"
)

type CompareFunc func([]*FTSEnt, []*FTSEnt) int

// Taken from gnulib's "fts_.h" (which is a superset of Linux's
// <fts.h>) in order to keep from using CGO (even for constants).
// curl --silent https://raw.githubusercontent.com/coreutils/gnulib/master/lib/fts_.h | grep -e "# define" | sed 's/# define//g' >> decls.txt
const (
	FTS_H = 1

	FTS_COMFOLLOW = 1 << iota
	FTS_LOGICAL
	FTS_NOCHDIR
	FTS_NOSTAT
	FTS_PHYSICAL
	FTS_SEEDOT
	FTS_XDEV
	FTS_WHITEOUT
	FTS_TIGHT_CYCLE_CHECK
	FTS_CWDFD
	FTS_DEFER_STAT
	FTS_NOATIME
	FTS_VERBATIM

	FTS_OPTIONMASK = 0x1fff // 8191
	FTS_NAMEONLY   = 0x2000 // 8192
	FTS_STOP       = 0x4000 // 16384

	FTS_ROOTPARENTLEVEL = iota - 1
	FTS_ROOTLEVEL
	FTS_D
	FTS_DC
	FTS_DEFAULT
	FTS_DNR
	FTS_DOT
	FTS_DP
	FTS_ERR
	FTS_F
	FTS_INIT
	FTS_NS
	FTS_NSOK
	FTS_SL
	FTS_SLNONE
	FTS_W

	FTS_DONTCHDIR = 0x01
	FTS_SYMFOLLOW = 0x02

	FTS_AGAIN = iota + 1
	FTS_FOLLOW
	FTS_NOINSTR
	FTS_SKIP
)

type FTSEnt struct {
	cycle         *FTSEnt
	parent        *FTSEnt
	link          *FTSEnt
	dirp          *dirent.Stream
	number        int64
	ptr           uintptr
	accPath       string
	path          string
	errno         int
	symFd         int
	pathLen       int // len(path)
	fts           *FTS
	level         int    // ptrdiff_t is +- a word in my stdint.h
	nameLen       string // len(name)
	dirsRemaining uint   // nlink_t is either a ulong or uword
	info          uint8
	flags         uint8
	instr         uint8
	stat          *syscall.Stat_t
	name          string
}

type FTS struct {
	cur          *FTSEnt
	child        *FTSEnt
	array        []*FTSEnt
	dev          uint64
	path         string
	rfd          int
	cwdFd        int
	pathLen      int // len(path)
	numItems     int // len(array)
	compare      CompareFunc
	opts         int
	leafOptWorks map[uint64]bool
	cycle        interface{} // either ADMap or *cycle.State
	ftsFdRing    *rb.Buffer
}

func (f *FTS) hasCycleAndLogicalOpts() bool {
	return f.opts&(FTS_TIGHT_CYCLE_CHECK|FTS_LOGICAL) != 0
}

const (
	DT_BLK = iota + 1
	DT_CHR
	DT_DIR
	DT_FIFO
	DT_LNK
	DT_REG
	DT_SOCK
)
