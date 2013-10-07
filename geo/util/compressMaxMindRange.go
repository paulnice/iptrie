// Copyright 2013 The iptrie Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

func uint32ToBytes(n uint32) []byte {
	t := n
	a := make([]byte, 4)
	a[3] = byte(t & 0x000000FF)
	t = t >> 8
	a[2] = byte(t & 0x000000FF)
	t = t >> 8
	a[1] = byte(t & 0x000000FF)
	t = t >> 8
	a[0] = byte(t & 0x000000FF)
	return a
}

// CompressRangeLocation compresses the file obtained from updateMaxMindRange
// and generates a binary and ascii file.
func CompressRangeLocation() error {

	filenamer := "maxmind_range_ipv4_location.csv"
	filenamew1 := "maxmind_range_ipv4_location_compressed.csv"
	filenamew2 := "maxmind_range_ipv4_location_compressed.bin"
	f1, errr := os.Open(filenamer)
	f2, errw2 := os.Create(filenamew1)
	f3, errw3 := os.Create(filenamew2)
	defer f1.Close()
	defer f2.Close()
	defer f3.Close()
	if errr != nil {
		panic(errr)
	}
	if errw2 != nil {
		panic(errw2)
	}
	if errw3 != nil {
		panic(errw3)
	}

	r := csv.NewReader(f1)
	var record, recordNext []string
	var start, end, locId, startNext, endNext, locIdNext uint32
	var num uint64
	record, errr = r.Read()
	if errr == io.EOF || len(record) == 0 {
		err := fmt.Errorf("File is empty/First line cannot be blank\n")
		return err
	}

	num, _ = strconv.ParseUint(record[0], 10, 32)
	start = uint32(num)
	num, _ = strconv.ParseUint(record[1], 10, 32)
	end = uint32(num)
	num, _ = strconv.ParseUint(record[2], 10, 32)
	locId = uint32(num)

	for {
		recordNext, errr = r.Read()
		if errr == io.EOF {
			break
		}
		if len(recordNext) == 0 {
			continue
		}

		num, _ = strconv.ParseUint(recordNext[0], 10, 32)
		startNext = uint32(num)
		num, _ = strconv.ParseUint(recordNext[1], 10, 32)
		endNext = uint32(num)
		num, _ = strconv.ParseUint(recordNext[2], 10, 32)
		locIdNext = uint32(num)

		if locId != locIdNext || end+1 != startNext {
			fmt.Fprintf(f2, "%d,%d,%d\n", start, end, locId)
			f3.Write(uint32ToBytes(start))
			f3.Write(uint32ToBytes(end))
			f3.Write(uint32ToBytes(locId))
			start = startNext
			end = endNext
			locId = locIdNext
		} else {
			// if lodid == locIdNext and end+1 == startNext then compress
			// don't update start
			num, _ = strconv.ParseUint(recordNext[1], 10, 32)
			end = uint32(num)
			// locIdNext is the same locId
		}
	}
	// writting the last record in the file
	fmt.Fprintf(f2, "%d,%d,%d\n", start, end, locId)
	f3.Write(uint32ToBytes(start))
	f3.Write(uint32ToBytes(end))
	f3.Write(uint32ToBytes(locId))
	bufend := new(bytes.Buffer)
	errw2 = binary.Write(bufend, binary.LittleEndian, '\n')
	if errw2 != nil {
		panic(errw2)
	}
	f3.Write(bufend.Bytes())
	return nil
}
