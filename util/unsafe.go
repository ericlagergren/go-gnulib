package util

import "unsafe"

func i8tob(a []int8) []byte {
	return *(*[]byte)(unsafe.Pointer(&a[0]))
}

func btoi8(a []byte) []int8 {
	return *(*[]int8)(unsafe.Pointer(&a[0]))
}

func i8toStr(a []int8) string {
	return string(i8tob(a))
}
