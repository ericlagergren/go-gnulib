### Go-Utmp

This is a collection of most of the assorted Linux/POSIX-esque UTMP
and WTMP functions.

It includes FreeBSD.

Thankfully, since Go doesn't have to deal with this stupid backwards
compatibility, I fiddled a bit with the API and instead of sharing a fd
and having idiotic `setutent()` or `endutent()` for absolutely zero reason,
most `utent` functions simply pass a pointer to an open file.

The code is licensed under the LGPLv3 and GPLv2 (per-file basis).