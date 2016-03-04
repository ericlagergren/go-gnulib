// +build !linux,!windows

package sysinfo

func PhysmemAvailable() int64 {
	return 1e9
}
