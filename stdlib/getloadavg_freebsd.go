package stdlib

import (
	"syscall"
	"unsafe"
)

type loadavg struct {
	ldavg [3]uint32
	scale uint64
}

// Get system load averages.
func GetLoadAvg(avg *[3]float64) int {
	v, err := syscall.Sysctl("vm.loadavg")
	if err != nil {
		return -1
	}

	b := []byte(v)
	var l loadavg = *(*loadavg)(unsafe.Pointer(&b[0]))

	scale := float64(l.scale)

	i := 0
	for ; i < 3; i++ {
		avg[i] = float64(l.ldavg[i]) / scale
	}

	return i
}
