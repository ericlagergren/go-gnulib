package sysinfo

import (
	"syscall"
)

// Returns available physical memory
func PhysmemAvail() uint64 {
	info := syscall.Sysinfo_t{}
	err := syscall.Sysinfo(&info)
	if err != nil {
		return 0
	}
	return info.Freeram
}
