package fts

import (
	"github.com/EricLagergren/go-gnulib/cycle"
)

type ActiveDir struct {
	dev uint64
	ino uint64
	ent *FTSEnt
}

func (a *ActiveDir) key() ADKey {
	return ADKey{dev: a.dev, ino: a.ino}
}

// ADKey is used to simulate GNU's hash.c hashmap.
// With GNU's hashmap, you specify hashing and comparing
// functions to, respectively, hash your key and compare
// for equality. Unfortunately, this is impossible to do
// with Go's map types. Since GNU's comparator function
// simply checks the device and inode values for equality,
// we use a key with only those two values and use the
// *FTSEnt as its value.
type ADKey struct{ dev, ino uint64 }

type ADMap map[ADKey]*ActiveDir

// Returns an ActiveDir with its contents initialized to the
// contents of the FTSent structure.
func (f *FTSEnt) NewActiveDir() *ActiveDir {
	return &ActiveDir{
		dev: f.stat.Dev,
		ino: f.stat.Ino,
		ent: f,
	}
}

func (a *ActiveDir) SameDir(b *ActiveDir) bool {
	return a.dev == b.dev && a.ino == b.ino
}

// SetupDir initializes our FTS.
func (f *FTS) SetupDir() {
	if f.hasCycleAndLogicalOpts() {
		f.cycle = make(ADMap, 31)
	} else {
		f.cycle = cycle.NewState()
	}
}

// EnterDir enters a directory during a file tree walk.
func (f *FTS) EnterDir(ent *FTSEnt) {
	if f.hasCycleAndLogicalOpts() {

		// Three cheers for a high-level language's abstractions.
		ad := ent.NewActiveDir()

		if fromTable, ok := f.cycle.(ADMap)[ad.key()]; !ok {
			ent.cycle = fromTable.ent
			ent.info = FTS_DC
		} else {
			f.cycle.(ADMap)[ad.key()] = ad
		}
	} else {
		if f.cycle.(*cycle.State).IsCycle(ent.stat) {

			ent.cycle = ent
			ent.info = FTS_DC
		}
	}
}

func (f *FTS) LeaveDir(ent *FTSEnt) {
	if f.hasCycleAndLogicalOpts() {
		delete(f.cycle.(ADMap), ADKey{
			dev: ent.stat.Dev,
			ino: ent.stat.Ino,
		})
	} else {
		if ent.parent != nil && 0 <= ent.parent.level {
			f.cycle.(*cycle.State).
				ChdirUp(ent.parent.stat, ent.stat)
		}
	}
}
