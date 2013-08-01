// Copyright 2013 The iptrie Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package locmap

// Data has the necessary fields needed for adding information to the locationMap
type Data struct {
	Status     bool
	ResourceId string
	ServerId   string
	Lat        float64
	Lon        float64
}
