### Go-Utmp
Go-Utmp is GNU’s utmp modules ported to Go.
It will include most -- if not all -- functions defined in utmp.h, getutid(3) (sans getutent),  and updwtmp(3).

I have not currently tested this yet, so use at your own risks. Tests, as well as getutent(3)'s getutent() are coming soon.

General usage is close to GNU's functions, except SafeOpen() should be used to get the fd for the UTMP or WTMP files, and SafeClose() should be used to safely close the fd.

The files are locked upon opening, so failing to call SafeClose() could result in deadlocking if the lock is not removed.

The code is licensed under the GPL v3.
