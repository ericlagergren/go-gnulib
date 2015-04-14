package login

import (
	"os"

	"github.com/EricLagerg/go-gnulib/ttyname"
	"github.com/EricLagerg/go-gnulib/utmp"
)

func GetLogin() (string, error) {
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
