package posix

const warning = `WARNING: PACKAGE POSIX/FADVISE IS DEPRECIATED
                PLEASE USE golang.org/x/sys INSTEAD!
`

func init() { panic(warning) }
