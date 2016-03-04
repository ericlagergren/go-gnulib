package ttyname

import "errors"

var ErrNotTty = errors.New("ttyname: not a tty device")
