// Copyright 2013 The iptrie Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iptrie

import (
	"math/rand"
	"net"
	"testing"
)

const (
	seed            = 0
	BenchTrieSizeSm = 15000
	BenchTrieSizeMd = 200000
	BenchTrieSizeLg = 2500000
)

var (
	src = rand.NewSource(seed)
	rnd = rand.New(src)
)

// may not be a valid IPv4 address.
func rndIPv4() net.IP {
	return net.IPv4(
		byte(1+(rnd.Uint32()%253)),
		byte(1+(rnd.Uint32()%253)),
		byte(1+(rnd.Uint32()%253)),
		byte(1+(rnd.Uint32()%253)))
}

// may not be a valid IPv6 address.
func rndIPv6() net.IP {
	p := make(net.IP, net.IPv6len)
	for i := range p {
		p[i] = byte(1 + (rnd.Uint32() % 253))
	}
	return p
}

func buildIPTrie(size int, tt *IPTrie, ipv4, ipv6 bool) {
	var ip net.IP
	var s string
	var e string
	var j int
	var sz int
	if ipv4 && ipv6 {
		sz = size / 2
	} else {
		sz = size
	}
	for i := 0; i < sz; i++ {
		if ipv4 {
			ip = rndIPv4()
			ip[15] = byte(0)
			s = ip.String()
			ip[15] = byte(254)
			e = ip.String()
			tt.AddRange(s, e, nil)
		}
		if ipv6 {
			ip = rndIPv6()
			for j = 7; j < 16; j++ {
				ip[j] = byte(0)
			}
			s = ip.String()
			for j = 7; j < 16; j++ {
				ip[j] = byte(254)
			}
			e = ip.String()
			tt.AddRange(s, e, nil)
		}
	}
}

func isEmpty(t *IPTrie) bool {
	t.m.Lock()
	defer t.m.Unlock()
	return len(t.kids) == 0 && t.parent == nil
}

func hasAddr(t *IPTrie, addr net.IP) bool {
	p := t
	for i := range addr {
		p = p.kids[addr[i]]
		if p == nil {
			return false
		}
	}
	return true
}

func hasRange(t *IPTrie, addrStart, addrEnd net.IP) bool {
	ps := t
	for i := range addrStart {
		ps = ps.kids[addrStart[i]]
		if ps == nil {
			return false
		}
	}
	if ps.rangeStart == nil {
		return false
	}
	if ps.rangeStart != ps {
		return false
	}
	pe := t
	for i := range addrEnd {
		pe = pe.kids[addrEnd[i]]
		if pe == nil {
			return false
		}
	}
	if pe.rangeStart == nil {
		return false
	}
	if pe.rangeStart != ps {
		return false
	}
	return true
}

func TestRndIP(t *testing.T) {
	ipa := rndIPv4()
	ipb := rndIPv4()
	if ipa.String() == ipb.String() {
		t.Fail()
	}
	ipa = rndIPv6()
	ipb = rndIPv6()
	if ipa.String() == ipb.String() {
		t.Fail()
	}
}

func TestUint32ToIPv4(t *testing.T) {
	s := Uint32ToIPv4(3232246374).String()
	if s != "192.168.42.102" {
		t.Fail()
	}
}

func TestIPv4ToUInt32(t *testing.T) {
	n := IPv4ToUInt32(net.ParseIP("192.168.42.102"))
	if n != uint32(3232246374) {
		t.Fail()
	}
}

func TestNewIPTrie(t *testing.T) {
	tt := NewIPTrie()
	if tt.kids == nil {
		t.Fail()
	}
	if tt.m == nil {
		t.Fail()
	}
	if !isEmpty(tt) {
		t.Fail()
	}
}

func TestAdd(t *testing.T) {
	tt := NewIPTrie()
	a := "192.168.42.102"
	addr := net.ParseIP(a)
	tt.Add(a, nil)
	if !hasAddr(tt, addr) {
		t.Fail()
	}
	if hasAddr(tt, net.ParseIP("192.168.42.103")) {
		t.Fail()
	}
}

func TestAddNum(t *testing.T) {
	tt := NewIPTrie()
	tt.AddNum(3232246374, nil)
	addr := net.ParseIP("192.168.42.102")
	if !hasAddr(tt, addr) {
		t.Fail()
	}
}

func TestAddRange(t *testing.T) {
	tt := NewIPTrie()
	s := "192.168.42.1"
	e := "192.168.42.254"
	sAddr := net.ParseIP(s)
	eAddr := net.ParseIP(e)
	tt.AddRange(s, e, nil)
	if !hasRange(tt, sAddr, eAddr) {
		t.Error("range")
	}
}

func TestAddRangeNum(t *testing.T) {
	tt := NewIPTrie()
	sAddr := net.ParseIP("192.168.42.1")
	eAddr := net.ParseIP("192.168.42.254")
	tt.AddRangeNum(3232246273, 3232246526, nil)
	if !hasRange(tt, sAddr, eAddr) {
		t.Error("range")
	}
}

func TestAddRangeIp(t *testing.T) {
	tt := NewIPTrie()
	sAddr := net.ParseIP("192.168.42.1")
	eAddr := net.ParseIP("192.168.42.254")
	sBytes := Uint32ToIPv4(3232246273).To16()
	eBytes := Uint32ToIPv4(3232246526).To16()
	tt.AddRangeIp(sBytes, eBytes, nil)
	if !hasRange(tt, sAddr, eAddr) {
		t.Error("range")
	}
}

type testData struct {
	i int
}

func TestGet(t *testing.T) {
	tt := NewIPTrie()
	a := "192.168.31.102"
	da := &testData{10}
	tt.Add(a, da)
	s := "192.168.42.1"
	e := "192.168.42.254"
	db := &testData{20}
	tt.AddRange(s, e, db)
	tr := tt.Get(a).(*testData)
	if tr == nil {
		t.Error("1")
	}
	if tr.i != 10 {
		t.Error("2")
	}
	tr = tt.Get("192.168.42.102").(*testData)
	if tr == nil {
		t.Error("3")
	}
	if tr.i != 20 {
		t.Error("4")
	}
	tn := tt.Get("192.168.43.1")
	if tn != nil {
		t.Error("5")
	}
}

func BenchmarkRndIPv4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = rndIPv4().String()
	}
}

func BenchmarkRndIPv6(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = rndIPv6().String()
	}
}

func BenchmarkUint32ToIPv4(b *testing.B) {
	n := IPv4ToUInt32(net.ParseIP("1.0.0.1"))
	for i := 0; i < b.N; i++ {
		_ = Uint32ToIPv4(n)
	}
}

func BenchmarkIPv4ToUInt32(b *testing.B) {
	a := net.ParseIP("192.168.42.1")
	for i := 0; i < b.N; i++ {
		_ = IPv4ToUInt32(a)
	}
}

func BenchmarkNewIPTrie(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewIPTrie()
	}
}

func BenchmarkAddIPv4(b *testing.B) {
	tt := NewIPTrie()
	var a string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		a = rndIPv4().String()
		b.StartTimer()
		tt.Add(a, nil)
	}
}

func BenchmarkAddIPv6(b *testing.B) {
	tt := NewIPTrie()
	var a string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		a = rndIPv6().String()
		b.StartTimer()
		tt.Add(a, nil)
	}
}

func BenchmarkAddNum(b *testing.B) {
	tt := NewIPTrie()
	var a uint32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		a = IPv4ToUInt32(rndIPv4())
		b.StartTimer()
		tt.AddNum(a, nil)
	}
}

func BenchmarkAddRangeIPv4(b *testing.B) {
	tt := NewIPTrie()
	var ip net.IP
	var s string
	var e string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ip = rndIPv4()
		ip[15] = byte(0)
		s = ip.String()
		ip[15] = byte(254)
		e = ip.String()
		b.StartTimer()
		tt.AddRange(s, e, nil)
	}
}

func BenchmarkAddRangeIPv6(b *testing.B) {
	tt := NewIPTrie()
	var ip net.IP
	var s string
	var e string
	var j int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ip = rndIPv6()
		for j = 7; j < 16; j++ {
			ip[j] = byte(0)
		}
		s = ip.String()
		for j = 7; j < 16; j++ {
			ip[j] = byte(254)
		}
		e = ip.String()
		b.StartTimer()
		tt.AddRange(s, e, nil)
	}
}

func BenchmarkAddRangeNum(b *testing.B) {
	tt := NewIPTrie()
	var ip net.IP
	var s uint32
	var e uint32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ip = rndIPv4()
		ip[15] = byte(0)
		s = IPv4ToUInt32(ip)
		ip[15] = byte(254)
		e = IPv4ToUInt32(ip)
		b.StartTimer()
		tt.AddRangeNum(s, e, nil)
	}
}

func BenchmarkGetIPv4(b *testing.B) {
	tt := NewIPTrie()
	buildIPTrie(BenchTrieSizeSm, tt, true, false)
	var ip net.IP
	var s string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ip = rndIPv4()
		s = ip.String()
		b.StartTimer()
		_ = tt.Get(s)
	}
}

func BenchmarkGetIPv6(b *testing.B) {
	tt := NewIPTrie()
	buildIPTrie(BenchTrieSizeSm, tt, false, true)
	var ip net.IP
	var s string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ip = rndIPv6()
		s = ip.String()
		b.StartTimer()
		_ = tt.Get(s)
	}
}

func BenchmarkGetSm(b *testing.B) {
	tt := NewIPTrie()
	buildIPTrie(BenchTrieSizeSm, tt, true, true)
	var ip net.IP
	var s string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ip = rndIPv4()
		s = ip.String()
		b.StartTimer()
		_ = tt.Get(s)
	}
}

func BenchmarkGetMd(b *testing.B) {
	tt := NewIPTrie()
	buildIPTrie(BenchTrieSizeSm, tt, true, true)
	var ip net.IP
	var s string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ip = rndIPv4()
		s = ip.String()
		b.StartTimer()
		_ = tt.Get(s)
	}
}

func BenchmarkGetLg(b *testing.B) {
	tt := NewIPTrie()
	buildIPTrie(BenchTrieSizeSm, tt, true, true)
	var ip net.IP
	var s string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ip = rndIPv4()
		s = ip.String()
		b.StartTimer()
		_ = tt.Get(s)
	}
}
