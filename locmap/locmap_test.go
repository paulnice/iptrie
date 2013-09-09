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
	if err == nil {
		t.Errorf("GetServer: boundary check failed")
	}
	_, err = m.GetServer(-91, 179, "2")
	if err == nil {
		t.Errorf("GetServer: boundary check failed")
	}
}

func checkResources(m *LocationMap, t *testing.T) string {

	result := ""
	l, _ := m.GetServer(0, 0, "1")
	if l == "" {
		t.Errorf("checkResources:GetServer: nothing here")
	}
	result += l + ":"

	l, _ = m.GetServer(0, 0, "2")
	if l == "" {
		t.Errorf("checkResources:GetServer: nothing here")
	}
	result += l + ":"

	l, _ = m.GetServer(0, 0, "3")
	if l == "" {
		t.Errorf("checkResouces:GetServer: nothing here")
	}
	result += l
	return result
}

func checkLocations(m *LocationMap, t *testing.T) string {

	result := ""
	l, _ := m.GetServer(1, 2, "1")
	if l == "" {
		t.Errorf("checkLocations:GetServer: nothing here")
	}
	result += l + ":"

	l, _ = m.GetServer(88, 172, "1")
	if l == "" {
		t.Errorf("checkLocations:GetServer: nothing here")
	}
	result += l + ":"

	l, _ = m.GetServer(88, 172, "2")
	if l == "" {
		t.Errorf("checkLocations:GetServer: nothing here")
	}
	result += l
	return result
}

func TestUpdate(t *testing.T) {

	memstats := &runtime.MemStats{}
	before := getMemStats(memstats)
	m := NewLocationMap()
	runtime.GC()
	after := getMemStats(memstats)
	fmt.Printf("New LocationMap: %fKB\n", math.Abs(float64(after-before)))

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
	fmt.Printf("Updated LocationMap: %fKB\n", math.Abs(float64(after-before)))

	for i := 0; i <= maxLat; i++ {
		for j := 0; j <= maxLon; j++ {
			location := m.mapp[i][j]
			if location == nil {
				t.Errorf("Update: init check failed")
			}
		}
	}
	result := checkResources(m, t)
	if result != "154.67.34.2:154.67.34.2:154.67.34.2" {
		t.Errorf("Update: init checkResources failed")
	}

	result = checkLocations(m, t)
	if result != "154.67.34.2:155.67.34.2:154.67.34.2" {
		t.Errorf("Update: init checkLocations failed")
	}

	m.Update(nil, allEntries)

	result = checkLocations(m, t)
	if result != "154.67.34.2:155.67.34.2:154.67.34.2" {
		t.Errorf("Update: nil checkLocations failed")
	}

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
			break
		}
	}
	m.Update(&e, allEntries)

	result = checkLocations(m, t)
	if result != "154.67.34.2:154.67.34.2:154.67.34.2" {
		t.Errorf("Update: change location check")
	}
}
