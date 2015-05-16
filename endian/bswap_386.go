package endian

func Bswap32(x uint32) uint32 {
	return ((x << 24) & 0xff000000) |
		((x << 8) & 0x00ff0000) |
		((x >> 8) & 0x0000ff00) |
		((x >> 24) & 0x000000ff)
}

func Bswap64(x uint64) uint64 {
	a := Bswap32(uint32((x & 0x00000000ffffffff)))
	b := Bswap32(uint32(((x >> 32) & 0x00000000ffffffff)))

	return uint64(a)<<32 | uint64(b)
}
