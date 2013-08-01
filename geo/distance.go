// Copyright 2013 The iptrie Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geo

import (
	"math"
)

const (
	EarthRadius    = 6371
	DubEarthRadius = 2 * EarthRadius
)

func deg2Rad(degree float64) float64 {
	return (degree * math.Pi) / 180
}

//Distance computes the distance between two geolocations
func Distance(lat1, lon1, lat2, lon2 float64) float64 {

	dlat := deg2Rad(lat2 - lat1)
	dlon := deg2Rad(lon2 - lon1)
	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(deg2Rad(lat1))*math.Cos(deg2Rad(lat2))*
			math.Sin(dlon/2)*math.Sin(dlon/2)
	return DubEarthRadius * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
