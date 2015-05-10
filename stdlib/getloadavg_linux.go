package stdlib

import (
	"bytes"
	"io"
	"os"
	"strconv"
)

// Put the 1 minute, 5 minute and 15 minute load averages
// into the first avg. Return the number written (never more than, but may
// be less than, 3), or -1 if an error occurred.
func GetLoadAvg(avg *[3]float64) int {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return -1
	}
	defer file.Close()

	var buf [65]byte

	n, err := file.Read(buf[:])
	if err != nil && err != io.EOF {
		return -1
	}

	i, j, end := 0, 0, 0
	for ; i < 3; i++ {
		end += bytes.IndexByte(buf[end:n], ' ')

		avg[i], err = strconv.ParseFloat(string(buf[j:end]), 64)
		if err != nil {
			return -1
		}

		// Skip over the space and mark our last position
		end++
		j = end
	}

	return i
}
