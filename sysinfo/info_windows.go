package sysinfo

// #include <stdio.h>
// #include <windows.h>
import "C"

import "unsafe"

func PhysmemAvailable() uint64 {
	var msex C.MEMORYSTATUSEX
	msex.dwLength = C.DWORD(unsafe.Sizeof(msex))

	/* Preferable */
	if C.GlobalMemoryStatusEx(&msex) != C.FALSE {
		return uint64(msex.ullAvailPhys)
	}

	/* Fallback because it's incorrect over 4GB */
	var ms C.MEMORYSTATUS
	C.GlobalMemoryStatus(&ms)
	return uint64(ms.dwAvailPhys)
}
