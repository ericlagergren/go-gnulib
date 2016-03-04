// +build !darwin,!darwin

package sysinfo

import (
	"fmt"
	"testing"
)

func TestPhysmemAvailable(t *testing.T) {
	fmt.Println(PhysmemAvailable(), "bytes")
}
