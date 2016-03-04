// Package cycle implements GNU's cycle-check.c
package cycle

import "os"

// I don't know what this does.
const magic = 9827862

type State struct {
	info    os.FileInfo
	counter uint64
	magic   int
}

// NewState returns a pointer to a State initialized
// with the magic value.
func NewState() *State { return &State{magic: magic} }

// ChdirUp moves up one directory in the hierarchy.
// In C it's a do-while macro that's called
// "CYCLE_CHECK_REFLECT_CHDIR_UP".
func (s *State) ChdirUp(dir, subdir os.FileInfo) {
	if s.counter == 0 {
		panic("You must call IsCycle at least once before calling this function.")
	}

	if os.SameFile(s.info, dir) {
		assign(subdir, s.info)
	}
}

func isZeroOrPowerOfTwo(x uint64) bool {
	return (x & (x - 1)) == 0
}

// IsCycle returns true if a cycle is found.
// Call this function once per chdir call.
func (s *State) IsCycle(sb os.FileInfo) bool {
	if s.counter > 0 && os.SameFile(sb, s.info) {
		return true
	}

	s.counter++
	if isZeroOrPowerOfTwo(s.counter) {

		// Theoretical overflow. 2**64 is a lot.
		// GNU's cycle-check.c says theoretical a bunch of times,
		// so it *must* be theoretical. I mean, assuming you're
		// recursing through 18,446,744,073,709,551,615
		// directories is highly unlikely.
		if s.counter == 0 {
			return true
		}
		assign(s.info, sb)
	}

	return false
}
