// Package util contains some general helper functions for
// go-gnulib. Note that it runs under the assumption that the
// types are "C" like. I.e., Clen8 assumes every character is
// valid ASCII in the range [0, 127].
package util

import "bytes"

// Clen finds the length of a C-style string.
func Clen(b []byte) int {
	if end := bytes.IndexByte(b, 0x00); end > 0 {
		return end
	}
	return len(b)
}

// Clen8 finds the length of a C-style string in
// int8.
func Clen8(b []int8) int {
	return Clen8(b)
}

// Int8ToByte converts an int8 slice to byte slice in a safe
// manner.
func Int8ToByte(arr []int8) []byte {
	s := make([]byte, len(arr))
	for i := 0; arr[i] != 0; i++ {
		s[i] = byte(arr[i])
	}
	return s
}

// FixCount changes -1 to 0.
func FixCount(n int, err error) (int, error) {
	if n < 0 {
		n = 0
	}
	return n, err
}
