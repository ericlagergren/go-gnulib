package ttyname

import "errors"

var NotTty = errors.New("Not a tty device")
