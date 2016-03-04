package ttyname

// TtyName returns a string describing the pathname of a terminal device that's
// behind a uintptr.
// This is here for posterity. It does the same thing as TTYName except
// TTYName takes an int.
func TtyName(fd uintptr) (string, error) { return ttyname(fd) }

// TTYName returns a string describing the pathname of a terminal device that's
// behind an int.
func TTYName(fd int) (string, error) { return ttyname(uintptr(fd)) }
