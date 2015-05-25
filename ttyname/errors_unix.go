package ttyname

import "errors"

var (
	ErrNotFound = errors.New("Device not found")
	ErrNotTty   = errors.New("Not a tty device")
)
