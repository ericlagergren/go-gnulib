package dirent

import "golang.org/x/sys/unix"

// isAbsent returns true if the file is absent in the directory.
func isAbsent(d *unix.Dirent) bool {
	return d.Fileno == 0
}
