package general

import "bytes"

// General helper functions

// Find length of a C-style string
func Clen(b []byte) int {
	end := bytes.IndexByte(b, '0')
	if end == -1 {
		return len(b)
	}

	// +1 because the index of an element in an array isn't the same as
	// the length from the beginning of the array to said element.
	return end + 1
}

// Convert an int8 slice to byte slice
// Usage bytes := Int8toByte(arr[:])
func Int8ToByte(arr []int8) []byte {
	s := make([]byte, len(arr))
	for i := 0; arr[i] != 0; i++ {
		s[i] = byte(arr[i])
	}
	return s[:]
}

// Change -1 to 0
func FixCount(n int, err error) (int, error) {
	if n < 0 {
		n = 0
	}
	return n, err
}
