// Copyright 2013 The iptrie Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geo

import (
	"bytes"
	"testing"
)

const blocks = `"3232235521","3232238335","A"
"3232253185","3232301055","B"
"3232301313","3232369407","C"`

const location = `"A","A","A","A","",-27.9864,26.7066,,
"B","B","B","B","",-17.8178,31.0447,,
"C","C","C","C","",-22.5700,17.0836,,`

func TestMaxmindIPv4(t *testing.T) {
	fBlock := bytes.NewBufferString(blocks)
	fLocation := bytes.NewBufferString(location)
	ipt := NewIPTrie()
	err := AddMaxmindIPv4City(ipt, fBlock, fLocation)
	if err != nil {
		t.Fatalf("TestMaxmindIPv4:AddMaxmindIPv4City: %v", err)
	}

	s := "192.168.8.8"
	i := ipt.Get(s)
	if i == nil {
		t.Errorf("%s Loc = %v\n", s, i)
	}
	s = "192.168.10.1"
	i = ipt.Get(s)
	if i == nil {
		t.Errorf("%s Loc = %v\n", s, i)
	}
	s = "192.170.2.53"
	i = ipt.Get(s)
	if i == nil {
		t.Errorf("%s Loc = %v\n", s, i)
	}
	s = "192.168.80.10"
	i = ipt.Get(s)
	if i == nil {
		t.Errorf("%s Loc = %v\n", s, i)
	}
	s = "192.169.0.24"
	i = ipt.Get(s)
	if i != nil {
		t.Errorf("%s Loc = %v\n", s, i)
	}
	s = "192.168.11.11"
	i = ipt.Get(s)
	if i != nil {
		t.Errorf("%s Loc = %v\n", s, i)
	}
}
