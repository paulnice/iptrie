// Copyright 2013 The iptrie Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package locmap

import (
	"container/list"
	"fmt"
	"math"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func getMemStats(m *runtime.MemStats) uint64 {
	runtime.ReadMemStats(m)
	return m.Alloc / 1024
}

func TestGetServer(t *testing.T) {

	m := NewLocationMap()
	_, err := m.GetServer(0, 400, "1")
	if err != nil {
		fmt.Printf("GetServer: %v\n", err)
	} else {
		t.Errorf("GetServer: boundary check failed")
	}
	_, err = m.GetServer(-91, 179, "2")
	if err != nil {
		fmt.Printf("GetServer: %v\n", err)
	} else {
		t.Errorf("GetServer: boundary check failed")
	}
}

func checkResources(m *LocationMap, t *testing.T) {
	fmt.Printf("Update: init resource check\n")
	l, _ := m.GetServer(0, 0, "1")
	if l == "" {
		t.Errorf("Nothing here")
	} else {
		fmt.Printf("Found %s\n", l)
	}

	l, _ = m.GetServer(0, 0, "2")
	if l == "" {
		t.Errorf("Nothing here")
	} else {
		fmt.Printf("Found %s\n", l)
	}

	l, _ = m.GetServer(0, 0, "3")
	if l == "" {
		t.Errorf("Nothing here")
	} else {
		fmt.Printf("Found %s\n", l)
	}
}

func checkLocations(m *LocationMap, check string, t *testing.T) {
	fmt.Printf("%s\n", check)
	l, _ := m.GetServer(1, 2, "1")
	if l == "" {
		t.Errorf("Nothing here")
	} else {
		fmt.Printf("Found %s\n", l)
	}
	l, _ = m.GetServer(88, 172, "1")
	if l == "" {
		t.Errorf("Nothing here")
	} else {
		fmt.Printf("Found %s\n", l)
	}
	l, _ = m.GetServer(88, 172, "2")
	if l == "" {
		t.Errorf("Nothing here")
	} else {
		fmt.Printf("Found %s\n", l)
	}
}

func TestUpdate(t *testing.T) {

	memstats := &runtime.MemStats{}
	before := getMemStats(memstats)
	m := NewLocationMap()
	runtime.GC()
	after := getMemStats(memstats)
	fmt.Printf("%fKB\n", math.Abs(float64(after-before)))

	allEntries := list.New()
	var testValues = []string{
		"true,1,154.67.34.2,-89,90",
		"true,2,154.67.34.2,-89,90",
		"true,3,154.67.34.2,-89,90",
		"true,1,155.67.34.2,88,172",
		"true,3,155.67.34.2,88,172",
	}

	for _, tt := range testValues {
		arr := strings.Split(tt, ",")
		present, _ := strconv.ParseBool(arr[0])
		lat, _ := strconv.ParseFloat(arr[3], 64)
		lon, _ := strconv.ParseFloat(arr[4], 64)
		e := Data{
			Status:     present,
			ResourceId: arr[1],
			ServerId:   arr[2],
			Lat:        lat,
			Lon:        lon,
		}
		allEntries.PushBack(e)
	}

	memstats = &runtime.MemStats{}
	before = getMemStats(memstats)
	for f := allEntries.Front(); f != nil; f = f.Next() {
		ff := f.Value.(Data)
		m.Update(&ff, nil)
	}
	runtime.GC()
	after = getMemStats(memstats)
	fmt.Printf("%fKB\n", math.Abs(float64(after-before)))

	for i := 0; i <= HiLat*2; i++ {
		for j := 0; j <= HiLon*2; j++ {
			location := m.mapp[i][j]
			if location == nil {
				t.Errorf("Update: init check failed")
			}
		}
	}
	checkResources(m, t)
	checkLocations(m, "Update: init location check", t)

	m.Update(nil, allEntries)
	checkLocations(m, "Update: nil location check", t)

	e := Data{
		Status:     false,
		ResourceId: "1",
		ServerId:   "155.67.34.2",
		Lat:        88,
		Lon:        172,
	}
	allEntries.PushBack(e)
	for f := allEntries.Front(); f != nil; f = f.Next() {
		ff := f.Value.(Data)
		if ff.ResourceId == "1" && ff.ServerId == "155.67.34.2" && ff.Status {
			allEntries.Remove(f)
		}
	}
	m.Update(&e, allEntries)
	checkLocations(m, "Update: change location check", t)

}
