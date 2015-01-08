package gnulib

// General helper functions that aren't exported but here for your
// copy and paste usage

// Find length of a C-style string
func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}

// clen() but for int8 instead of []byte
func clen8(n [256]int8) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}

// Convert an int8 array to byte slice
func int8ToByte(rel [256]int8) []byte {
	s := [256]byte{}
	for i := 0; i < len(rel); i++ {
		s[i] = byte(rel[i])
	}
	return s[:clen8(rel)]
}

// change -1 to 0
func fixCount(n int, err error) (int, error) {
	if n < 0 {
		n = 0
	}
	return n, err
}
