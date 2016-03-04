// Package ifdef implements some constants that don't exist on certain
// platforms (or have different names) that we can unify under a common
// name. For example, some Linux and Solaris has O_SEARCH, but
// it's not in Go's unix/syscall packages. Since O_SEARCH is preferable to
// O_RDONLY (works nearly the same on platforms without O_SEARCH), we
// use O_SEARCH which is defined as O_RDONLY on the platforms without
// O_SEARCH.
package ifdef
