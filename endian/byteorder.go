package math

import "unsafe"

var ByteOrder = -1

const (
	BigEndian = iota
	LittleEndian
)

func Htobe32(x uint32) uint32 {
	if ByteOrder == BigEndian {
		return x
	}
	return Bswap32(x)
}

func Htobe64(x uint64) uint64 {
	if ByteOrder == BigEndian {
		return x
	}
	return Bswap64(x)
}

func init() {
	var x uint32 = 0x01020304
	switch *(*byte)(unsafe.Pointer(&x)) {
	case 0x01: // Big Endian
		ByteOrder = BigEndian
	case 0x04: // Little Endian
		ByteOrder = LittleEndian
	default:
		panic("unknown endianness")
	}
}
