// Copyright 2013 The iptrie Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The locmap package provides a data structure that returns the
// serverId of the closest server for a given geolocation. It can
// handler mutiple readers and one writter.
//
package locmap

import (
	"container/list"
	"errors"
	"math"
	"sync"
	"time"

	"code.google.com/p/iptrie/geo"
)

type location struct {
	valMap  map[string]string  // ResourceId --> ServerId <could be ipaddress> of closest server
	distMap map[string]float64 // ResourceId --> distance to closest server
}

type LocationMap struct {
	mapp    [][]*location
	rwmutex *sync.RWMutex
	mutex   *sync.Mutex
	dataExp time.Time
}

const (
	maxLat = 180
	maxLon = 360
)

var (
	LoLat       float64 = -90
	HiLat       float64 = 90
	LoLon       float64 = -180
	HiLon       float64 = 180
	BadLocation         = errors.New("Geolocation value is out of range")
)

// NewLocationMap creates a new location map
func NewLocationMap() *LocationMap {

	c := make([][]*location, maxLat+1)
	for i := range c {
		c[i] = make([]*location, maxLon+1)
		for j := range c[i] {
			c[i][j] = &location{
				valMap:  make(map[string]string),
				distMap: make(map[string]float64),
			}
		}
	}
	return &LocationMap{
		mapp:    c,
		rwmutex: &sync.RWMutex{},
		mutex:   &sync.Mutex{},
		dataExp: time.Now(),
	}
}

// deepCopy creates a new location map with the contents of another location map
func (m *LocationMap) deepCopy() *LocationMap {
	m.rwmutex.RLock()
	mm := m.mapp
	m.rwmutex.RUnlock()

	c := make([][]*location, maxLat+1)
	for i := range c {
		c[i] = make([]*location, maxLon+1)
		for j := range c[i] {
			c[i][j] = mm[i][j]
		}
	}
	return &LocationMap{
		mapp: c,
	}
}

func (m *LocationMap) set(lat, lon int, data *location) *location {

	if data == nil {
		m.mapp[lat][lon] = &location{
			valMap:  make(map[string]string),
			distMap: make(map[string]float64),
		}
	} else {
		m.mapp[lat][lon] = data
	}
	return m.mapp[lat][lon]
}

// GetServer returns the serverId of the closest server for a given geolocation and resourceId.
func (m *LocationMap) GetServer(lat, lon float64, resourceId string) (string, error) {

	if lat < LoLat || lat > HiLat || lon < LoLon || lon > HiLon {
		return "", BadLocation
	}
	m.rwmutex.RLock()
	location := m.mapp[int(lat+maxLat/2)][int(lon+maxLon/2)]
	m.rwmutex.RUnlock()
	if location == nil {
		return "", nil
	}
	return location.valMap[resourceId], nil
}

func (m *LocationMap) setServer(lat, lon int, resourceId, serverId string, dist float64) {

	location := m.mapp[lat][lon]
	if location == nil {
		location = m.set(lat, lon, nil)
	}

	location.valMap[resourceId] = serverId
	location.distMap[resourceId] = dist
}

// Update LocationMap. Can have mutiple readers and one writter.
func (m *LocationMap) Update(e *Data, allEntries *list.List) {

	if e == nil {
		return
	}

	m.mutex.Lock()
	newMap := m.deepCopy() // readers can continue to read from m
	if e.Status {
		newMap.addServer(e)
	} else {
		newMap.removeServer(e, allEntries)
	}

	m.rwmutex.Lock()
	m.mapp = newMap.mapp
	m.rwmutex.Unlock()
	m.mutex.Unlock()
}

// UpdateMulti LocationMap.
func (m *LocationMap) UpdateMulti(updateEntries, allEntries *list.List) {

	if updateEntries.Front() == nil {
		return
	}

	m.mutex.Lock()
	newMap := m.deepCopy()
	for e := updateEntries.Front(); e != nil; e = e.Next() {
		ee := e.Value.(*Data)
		if ee.Status {
			newMap.addServer(ee)
		} else {
			newMap.removeServer(ee, allEntries)
		}
	}

	m.rwmutex.Lock()
	m.mapp = newMap.mapp
	m.rwmutex.Unlock()
	m.mutex.Unlock()
}

// addServer adds a online server or a new server
func (m *LocationMap) addServer(e *Data) {
	for i := 0; i <= maxLat; i++ {
		for j := 0; j <= maxLon; j++ {

			location := m.mapp[i][j]
			if location == nil {
				location = m.set(i, j, nil)
			}

			distMap := location.distMap
			currDist, ok := distMap[e.ResourceId]
			if !ok {
				currDist = math.MaxFloat64
			}
			newDist := geo.Distance(float64(i-maxLat/2), float64(j-maxLon/2), e.Lat, e.Lon)
			if newDist < currDist {
				m.setServer(i, j, e.ResourceId, e.ServerId, newDist)
			}
		}
	}
}

// removeServer removes and if possible replaces an offline or retired server
func (m *LocationMap) removeServer(e *Data, allEntries *list.List) {
	for i := 0; i <= maxLat; i++ {
		for j := 0; j <= maxLon; j++ {

			location := m.mapp[i][j]
			if location == nil {
				location = m.set(i, j, nil)
			}

			valMap := location.valMap
			if valMap[e.ResourceId] == e.ServerId {
				serverId := ""
				distMin := math.MaxFloat64
				for f := allEntries.Front(); f != nil; f = f.Next() {
					ff := f.Value.(Data)
					if ff.ResourceId == e.ResourceId && ff.Status {
						newDist := geo.Distance(float64(i-maxLat/2), float64(j-maxLon/2), ff.Lat, ff.Lon)
						if newDist < distMin {
							serverId = ff.ServerId
							distMin = newDist
						}
					}
				}
				m.setServer(i, j, e.ResourceId, serverId, distMin) //remove/replace
			}
		}
	}
}
