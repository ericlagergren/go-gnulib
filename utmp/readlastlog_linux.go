// Copyright (c) 2015 Eric Lagergren
// Use of this source code is governed by the LGPL 2.1 or later.

// This file implements reading of the LastLog file.

package utmp

import (
	"encoding/binary"
	"io"
)

// ReadLastLog reads n entries from a lastlog file.
// It returns an error if any reads fail without io.EOF.
// If n is less than 0 it will read the entire file.
func ReadLastLog(n int64) (logs []LastLog, err error) {

	file, err := Open(LastLogFile, Reading)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var l LastLog
	for i := int64(0); i != n; i++ {
		err := binary.Read(file, binary.LittleEndian, &l)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, err
}
