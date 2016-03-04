package stdlib

import (
	"reflect"
	"syscall"
	"unsafe"
)

type loadavg struct {
	ldavg [3]uint32
	scale uint64
}

// Get system load averages.
func GetLoadAvg(avg *[3]float64) int {
	s, err := syscall.Sysctl("vm.loadavg")
	if err != nil {
		return -1
	}

	b := bytes(s)
	l := *(*loadavg)(unsafe.Pointer(&b[0]))

	for i := range avg {
		avg[i] = float64(l.ldavg[i] / l.scale)
	}
	return 3
}

func bytes(s string) []byte {
	ss := *(*reflect.StringHeader)(reflect.UnsafePointer(&s))
	ret := reflect.SliceHeader{
		Data: ss.Data,
		Len:  ss.Len,
		Cap:  ss.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&ret))
}
