package login

import (
	"os"

	"github.com/EricLagergren/go-gnulib/ttyname"
	"github.com/EricLagergren/go-gnulib/utmp"
)

func GetLogin() (string, error) {

	// GetLogin is based on stdin, so find the name of the current terminal
	// and return nothing if it doesn't exist. According to GNU, this is what
	// DEC Unix, SunOS, Solaris, and HP-UX all do.
	name, err := ttyname.TtyName(os.Stdin.Fd())
	if err != nil {
		return "", err
	}

	u := new(utmp.Utmp)
	_ = copy(u.Line[:], []byte(name[5:]))

	file, err := os.Open(utmp.UtmpFile)
	if err != nil {
		return "", err
	}

	line, err := u.GetUtLine(file)
	if err != nil {
		return "", err
	}

	return string(line.User[:]), nil
}
