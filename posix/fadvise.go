package posix

const (
	POSIX_FADV_NORMAL = iota
	POSIX_FADV_RANDOM
	POSIX_FADV_SEQUENTIAL
	POSIX_FADV_WILLNEED
	POSIX_FADV_DONTNEED
	POSIX_FADV_NOREUSE
)

// POSIX fadvise()
func Fadvise(file *os.File, off, length int, advice uint32) error {
	_, _, errno := syscall.Syscall6(syscall.SYS_FADVISE64,
		file.Fd(),
		uintptr(off),
		uintptr(length),
		uintptr(advice), 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}
