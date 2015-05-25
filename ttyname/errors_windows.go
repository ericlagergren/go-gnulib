package ttyname

import "errors"

var ErrNotTty = errors.New("Not a tty device")
