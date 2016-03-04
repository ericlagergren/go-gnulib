package main

import (
	"fmt"
	"sort"
	"syscall"
)

type fts struct {
	array ftsbuf
}

type ftsent struct {
	name string
	stat *syscall.Stat_t
	link *ftsent
}

func main() {
	var fb fts
	fb.array = make([]*ftsent, 5)

	var fsb [5]ftsent
	fsb[4] = ftsent{
		name: "four",
		link: nil,
		stat: &syscall.Stat_t{Ino: 4, Dev: 1},
	}
	fsb[3] = ftsent{
		name: "three",
		link: &fsb[4],
		stat: &syscall.Stat_t{Ino: 3, Dev: 1},
	}
	fsb[2] = ftsent{
		name: "two",
		link: &fsb[3],
		stat: &syscall.Stat_t{Ino: 2, Dev: 1},
	}
	fsb[1] = ftsent{
		name: "one",
		link: &fsb[2],
		stat: &syscall.Stat_t{Ino: 1, Dev: 1},
	}
	fsb[0] = ftsent{
		name: "zero",
		link: &fsb[1],
		stat: &syscall.Stat_t{Ino: 0, Dev: 1},
	}

	fmt.Printf("result: %s\n", doSort(&fb, &fsb[0], 5).name)
}

func doSort(sp *fts, head *ftsent, nitems int) *ftsent {
	ap := sp.array
	for i, p := 0, head; p != nil; p, i = p.link, i+1 {
		fmt.Printf("p: %s\n", p.name)
		ap[i] = p
	}

	fmt.Println("")

	sort.Sort(sp.array)

	// ap = sp.array

	var i int
	for i, head = 0, ap[0]; nitems-1 > 0; i, nitems = i+1, nitems-1 {
		fmt.Println(ap[i].name)
		ap[i].link = ap[1]
	}

	ap[0].link = nil

	return head
}

type ftsbuf []*ftsent

func (f ftsbuf) Len() int           { return len(f) }
func (f ftsbuf) Less(i, j int) bool { return compare(f[i], f[j]) < 0 }
func (f ftsbuf) Swap(i, j int) {
	tmp := f[i]
	f[i] = f[j]
	f[j] = tmp
}

func compare(a, b *ftsent) int {
	if a.stat.Ino < b.stat.Ino {
		return -1
	}

	if b.stat.Ino < a.stat.Ino {
		return 1
	}

	return 0
}
