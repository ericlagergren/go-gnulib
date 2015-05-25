package ttyname

import "errors"

var (
	NotFound = errors.New("Device not found")
	NotTty   = errors.New("Not a tty device")
)
