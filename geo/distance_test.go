// Copyright 2013 The iptrie Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geo

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

func TestDeg2Rad(t *testing.T) {
	var deg2RadTests = []struct {
		in  float64
		out float64
	}{
		{-148.045900, -2.5838884028043663e+00},
		{-174.720268, -3.0494439429041273e+00},
		{-43.808838, -7.6460845683771372e-01},
		{116.683156, 2.0365052512100750e+00},
		{-50.975000, -8.8968158339352454e-01},
		{-135.704457, -2.3684895758630824e+00},
		{-171.318610, -2.9900738145146382e+00},
		{-64.017259, -1.1173119544249543e+00},
		{48.479435, 8.4612575916689381e-01},
		{159.537615, 2.7844566700973661e+00},
	}

	for _, tt := range deg2RadTests {
		in := tt.in
		if out := deg2Rad(in); math.Abs(out-tt.out) > 0.0001 {
			t.Errorf("deg2Rad (%.6f) = %.16e, want %.16e", tt.in, out, tt.out)
		}
	}
}

func TestDistance(t *testing.T) {
	var distanceTests = []struct {
		in  string
		out float64
	}{
		{"-2.889553,-109.605666,79.694850,-118.501399", 9.1967718989509740e+03},
		{"47.151388,2.422298,-81.045511,-61.002100", 1.4744325297905116e+04},
		{"-41.289424,-147.556514,-44.756153,72.970537", 9.6311801756727109e+03},
		{"-80.395725,132.502421,18.424523,159.589369", 1.1100404758305427e+04},
		{"-61.361999,-37.849193,35.208151,104.001376", 1.6066065428631984e+04},
		{"-56.211936,-21.278493,-66.636987,102.374581", 5.5766928265408778e+03},
		{"85.962601,159.172637,88.281903,28.206413", 5.9200386317786626e+02},
		{"5.876585,42.169977,-86.568689,-112.121189", 1.1004503547326169e+04},
		{"20.460931,-166.403707,-30.781144,160.995010", 6.6775250693147318e+03},
		{"9.358486,-50.306052,-49.755767,-110.279907", 8.7581261995618061e+03},
	}

	for _, tt := range distanceTests {
		split := strings.Split(tt.in, ",")
		lat1, _ := strconv.ParseFloat(split[0], 64)
		lon1, _ := strconv.ParseFloat(split[1], 64)
		lat2, _ := strconv.ParseFloat(split[2], 64)
		lon2, _ := strconv.ParseFloat(split[3], 64)
		if out := Distance(lat1, lon1, lat2, lon2); math.Abs(out-tt.out) > 0.01 {
			t.Errorf("distance(%s) = %.16e, want %.16e", tt.in, out, tt.out)
		}
	}
}
