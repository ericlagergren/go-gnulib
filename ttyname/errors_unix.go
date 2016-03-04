// +build !windows

package ttyname

import "errors"

var (
	ErrNotFound = errors.New("ttyname: device not found")
	ErrNotTty   = errors.New("ttyname: not a tty device")
)
