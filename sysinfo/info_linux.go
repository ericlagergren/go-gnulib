package sysinfo

// #include <unistd.h>
// #include <sys/types.h>
import "C"

// same as C's physmem_available()
func PhysmemAvail() int64 {

	// type C.long
	pages := C.sysconf(C._SC_AVPHYS_PAGES) // available pages
	size := C.sysconf(C._SC_PAGESIZE)      // page size

	// sanity check
	if 0 <= pages && 0 <= size {
		return int64(pages * size)
	}

	// we overflowed or something went wrong
	return -1
}
