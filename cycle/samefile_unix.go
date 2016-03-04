// +build !windows

package cycle

import (
	"os"

	"golang.org/x/sys/unix"
)

func assign(dest, src os.FileInfo) {
	dest.Sys().(*unix.Stat_t).Ino = src.Sys().(*unix.Stat_t).Ino
	dest.Sys().(*unix.Stat_t).Dev = src.Sys().(*unix.Stat_t).Dev
}
