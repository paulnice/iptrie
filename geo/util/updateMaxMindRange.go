// Copyright 2013 The iptrie Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Input: GeoLiteCity-Location.csv, maxmind_city_ipv4_blocks.csv
// Output: maxmind_range_ipv4_location.csv
package util

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	maxLat = 180
	maxLon = 360
)

var (
	c = [][]string{}
)

func initc() {

	c = make([][]string, maxLat+1)
	for i := range c {
		c[i] = make([]string, maxLon+1)
		for j := range c[i] {
			c[i][j] = ""
		}
	}
}

func getCount() {

	count := 0
	for i := 0; i <= maxLat; i++ {
		for j := 0; j <= maxLon; j++ {
			if c[i][j] == "" {
				continue
			} else {
				count++
			}
		}
	}
	fmt.Printf("There are %d locations\n", count)
}

func getNewLocationData() error {

	f, err := os.Open("GeoLiteCity-Location.csv")
	defer f.Close()
	if err != nil {
		panic(err)
	}

	r := csv.NewReader(f)
	r.TrailingComma = true
	r.FieldsPerRecord = -1

	var record []string
	var locId string
	var myLat, myLon int
	var lat, lon float64
	n := 0

	for {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if len(record) == 0 {
			continue
		}

		locId = record[0]
		lat, err = strconv.ParseFloat(record[5], 64)
		if err != nil {
			err = fmt.Errorf("Latitude at line# %d is not valid \n", n+1)
			return err
		}
		myLat = int(lat + maxLat/2)
		lon, err = strconv.ParseFloat(record[6], 64)
		if err != nil {
			err = fmt.Errorf("Longitude at line# %d is not valid \n", n+1)
			return err
		}
		myLon = int(lon + maxLon/2)

		if myLat > maxLat || myLon > maxLon || myLat < 0 || myLon < 0 {
			err = fmt.Errorf("Geolocation at line# %d is out of range \n", n+1)
			return err
		}
		c[myLat][myLon] += locId + ","
		n++
	}
	fmt.Printf("Done reading %d lines\n", n)
	getCount()
	return nil
}

// UpdateRangeLocation parses maxMind ranges [start, end]:locationId and updates
// the locationId of the range to a value that represents its location.
func UpdateRangeLocation() error {

	initc()
	err := getNewLocationData()
	if err != nil {
		return err
	}
	filenamer := "maxmind_city_ipv4_blocks.csv"
	filenamew := "maxmind_range_ipv4_location.csv"
	f1, errr := os.Open(filenamer)
	f2, errw := os.Create(filenamew)
	defer f1.Close()
	defer f2.Close()
	if errr != nil {
		panic(errr)
	}
	if errw != nil {
		panic(errw)
	}

	r := csv.NewReader(f1)
	var record []string
	var start, end, locId string
	var added bool
	n := 0
	for {
		record, errr = r.Read()
		if errr == io.EOF {
			break
		}
		if len(record) == 0 {
			continue
		}

		start = record[0]
		end = record[1]
		locId = record[2]

		added = false
		for i := 0; i <= maxLat && !added; i++ {
			for j := 0; j <= maxLon && !added; j++ {
				if c[i][j] == "" {
					continue
				} else {
					split := strings.Split(c[i][j], ",")
					for _, s := range split {
						if s == "" {
							continue
						}
						if s == locId {
							fmt.Fprintf(f2, "%s,%s,%d\n", start, end, i*1000+j)
							added = true
							break
						}
					}
				}
			}
		}
		if !added {
			err = fmt.Errorf("Range at line %d not added: %d,%d", n+1, start, end)
			return err
		}
		n++
	}
	fmt.Printf("Done reading %d lines\n", n)
	return nil
}
