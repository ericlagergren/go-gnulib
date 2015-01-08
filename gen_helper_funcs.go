package gnulib

// General helper functions

// Find length of a C-style string
func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}

// Convert an sized int8 array to byte slice
// Usage bytes := int8ToByte(arr[:])
func int8ToByte(arr []int8) []byte {
	s := make([]byte, len(arr))
	for i := 0; arr[i] != 0; i++ {
		s[i] = byte(arr[i])
	}
	return s[:]
}

// change -1 to 0
func fixCount(n int, err error) (int, error) {
	if n < 0 {
		n = 0
	}
	return n, err
}
