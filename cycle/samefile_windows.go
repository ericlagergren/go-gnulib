package cycle

import (
	"os"
	"syscall"
)

func assign(dest, src os.FileInfo) {
	dest.Sys().(*syscall.Win32FileAttributeData).Ino = src.Sys().(*syscall.Win32FileAttributeData).VolumeSerialNumber
	dest.Sys().(*syscall.Win32FileAttributeData).Dev = src.Sys().(*syscall.Win32FileAttributeData).FileIndexHigh
	dest.Sys().(*syscall.Win32FileAttributeData).Dev = src.Sys().(*syscall.Win32FileAttributeData).FileIndexLow
}
