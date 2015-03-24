package posix

import (
	"os"
	"testing"
)

// A better test would be to run this as well
// strace -f go test 2>&1 | grep fadvise64
func TestFadvise(t *testing.T) {
	file, err := os.Open("fadvise_linux.go")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	err = Fadvise64(int(file.Fd()), 0, 0, FADVISE_NORMAL)
	if err != nil {
		t.Fatal(err)
	}

	err = Fadvise64(int(file.Fd()), 0, 0, FADVISE_RANDOM)
	if err != nil {
		t.Fatal(err)
	}

	err = Fadvise64(int(file.Fd()), 0, 0, FADVISE_SEQUENTIAL)
	if err != nil {
		t.Fatal(err)
	}

	err = Fadvise64(int(file.Fd()), 0, 0, FADVISE_WILLNEED)
	if err != nil {
		t.Fatal(err)
	}

	err = Fadvise64(int(file.Fd()), 0, 0, FADVISE_DONTNEED)
	if err != nil {
		t.Fatal(err)
	}

	err = Fadvise64(int(file.Fd()), 0, 0, FADVISE_NOREUSE)
	if err != nil {
		t.Fatal(err)
	}

	// catch stdout and grep for output...
}
