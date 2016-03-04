// +build !solaris

package ifdef

import "golang.org/x/sys/unix"

const O_SEARCH = unix.O_RDONLY
