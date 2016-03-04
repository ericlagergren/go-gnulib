package endian

import "unsafe"

var ByteOrder = -1

const (
	BigEndian = iota
	LittleEndian
)

func Htobe16(x uint16) uint16 {
	if ByteOrder == BigEndian {
		return x
	}
	return Bswap16(x)
}

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

func Htole16(x uint16) uint16 {
	if ByteOrder == LittleEndian {
		return x
	}
	return Bswap16(x)
}

func Htole32(x uint32) uint32 {
	if ByteOrder == LittleEndian {
		return x
	}
	return Bswap32(x)
}

func Htole64(x uint64) uint64 {
	if ByteOrder == LittleEndian {
		return x
	}
	return Bswap64(x)
}

func Be16toh(x uint16) uint16 {
	if ByteOrder == BigEndian {
		return x
	}
	return Bswap16(x)
}

func Be32toh(x uint32) uint32 {
	if ByteOrder == BigEndian {
		return x
	}
	return Bswap32(x)
}

func Be64toh(x uint64) uint64 {
	if ByteOrder == BigEndian {
		return x
	}
	return Bswap64(x)
}

func Le16toh(x uint16) uint16 {
	if ByteOrder == LittleEndian {
		return x
	}
	return Bswap16(x)
}

func Le32toh(x uint32) uint32 {
	if ByteOrder == LittleEndian {
		return x
	}
	return Bswap32(x)
}

func Le64toh(x uint64) uint64 {
	if ByteOrder == LittleEndian {
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
		// Also could be mixed/middle endian?
		panic("unknown endianness")
	}
}
