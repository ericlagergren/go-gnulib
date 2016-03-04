// +build !darwin,!windows

package sysinfo

// #include <unistd.h>
import "C"

func PhysmemAvailable() int64 {
	pages := C.sysconf(C._SC_AVPHYS_PAGES)
	pagesize := C.sysconf(C._SC_PAGESIZE)
	if 0 <= pages && 0 <= pagesize {
		return int64(pages * pagesize)
	}
	return PhysmemTotal() / 4
}

func PhysmemTotal() int64 {
	pages := C.sysconf(C._SC_PHYS_PAGES)
	pagesize := C.sysconf(C._SC_PAGESIZE)
	if 0 <= pages && 0 <= pagesize {
		return int64(pages * pagesize)
	}
	return 64 * 1024 * 1024
}
