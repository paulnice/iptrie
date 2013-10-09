// Copyright 2013 The iptrie Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The iptrie package provides a data structure that allows matching an IP
// address to a range.  That range can hold data.  One possible use for such a
// data structure is building a DNS server or looking up geolocation information
// based on IP address.  For the latter example please see the code in the
// accompanying geo package.
package iptrie

import (
	"net"
	"sort"
	"sync"
)

// An IPTrie is used for efficient range and prefix matching on IP addresses.
type IPTrie struct {
	parent     *IPTrie
	data       interface{}
	b          byte
	kids       map[byte]*IPTrie
	rangeStart *IPTrie
	m          *sync.Mutex
}

// NewIPTrie should be used to create an empty IPTrie.
func NewIPTrie() *IPTrie {
	return newIPTrie(nil)
}

// newIPTrie ensures that the synchronization fields are initialized.
func newIPTrie(p *IPTrie) *IPTrie {
	return &IPTrie{
		parent: p,
		kids:   make(map[byte]*IPTrie),
		m:      &sync.Mutex{},
	}
}

// Add places the IP address into the IPTrie and saves the associated data for
// later retrieval.
func (t *IPTrie) Add(addr string, data interface{}) {
	lts := t.addStr(addr)
	lts.data = data
	lts.rangeStart = lts
}

// AddNum places the IP address (converted from the uint32) into the IPTrie and
// saves the associated data for later retrieval.  This is a convenience method
// since maxmind uses uint32 for their ranges of IPv4 addresses.
func (t *IPTrie) AddNum(addr uint32, data interface{}) {
	lts := t.add(Uint32ToIPv4(addr).To16())
	lts.data = data
	lts.rangeStart = lts
}

// AddRange places the range of IP addressed into the IPTrie and saves the
// associated data for later retrieval.
func (t *IPTrie) AddRange(sAddr, eAddr string, data interface{}) {
	lts := t.addStr(sAddr)
	lts.rangeStart = lts
	lts.data = data
	lte := t.addStr(eAddr)
	lte.rangeStart = lts
}

// AddRangeNum places the range of  IP addressed (converted from the uint32)
// into the IPTrie and saves the associated data for later retrieval.  This is a
// convenience method since maxmind uses uint32 for their ranges of IPv4
// addresses.
func (t *IPTrie) AddRangeNum(sAddr, eAddr uint32, data interface{}) {
	lts := t.add(Uint32ToIPv4(sAddr).To16())
	lts.rangeStart = lts
	lts.data = data
	lte := t.add(Uint32ToIPv4(eAddr).To16())
	lte.rangeStart = lts
}

// AddRangeIp places the range of IP addresses proivded in their 16-byte
// representation into the IPTrie and saves the associated data for later
// retrieval.  This method is useful when the IP addresses are being read as an
// array of bytes say from a binary file.
func (t *IPTrie) AddRangeIp(sAddr, eAddr []byte, data interface{}) {
	lts := t.add(sAddr)
	lts.rangeStart = lts
	lts.data = data
	lte := t.add(eAddr)
	lte.rangeStart = lts
}

// AddCIDRRange computes a range of IP addresses from a vavid CIDR string and
// places the range into the IPTrie and saves the associated data for later
// retrieval.
func (t *IPTrie) AddCIDRRange(addr string, data interface{}) {
	sAddr, eAddr := cidrToRange(addr)
	if sAddr == nil || eAddr == nil {
		return
	}
	t.AddRangeIp(sAddr, eAddr, data)
}

func (t *IPTrie) addStr(addr string) *IPTrie {
	ip := net.ParseIP(addr)
	return t.add(ip.To16())
}

func (t *IPTrie) add(a []byte) *IPTrie {
	e := len(a) - 1
	for ; e > 0; e-- {
		if a[e] != 0 {
			break
		}
	}
	p := t
	var k *IPTrie
	for i := 0; i <= e; i++ {
		p.m.Lock()
		k = p.kids[a[i]]
		p.m.Unlock()
		if k == nil {
			k = newIPTrie(p)
			k.b = a[i]
			p.m.Lock()
			p.kids[a[i]] = k
			p.m.Unlock()
		}
		p = k
	}
	return p
}

// Get returns the data associated with the longest prefix or most compact
// range.  If the prefix length is zero or the address is outside of all ranges
// the result is nil.
func (t *IPTrie) Get(addr string) interface{} {
	ip := net.ParseIP(addr)
	k, i := t.get(ip.To16())
	if k.data != nil {
		return k.data
	}
	var b byte
	if i < len(ip) {
		b = ip[i]
	} else if k.parent == nil {
		return nil
	} else {
		b = k.b
		k = k.parent
	}
	var r *IPTrie
	for {
		r = k.getData(int(b))
		if r != nil {
			if r.data != nil {
				return r.data
			} else if r.rangeStart != nil {
				return nil
			}
		}
		if k.parent == nil {
			return nil
		}
		b = k.b
		k = k.parent
	}
	panic("Get")
}

func (t *IPTrie) get(a []byte) (*IPTrie, int) {
	p := t
	var k *IPTrie
	for i := range a {
		p.m.Lock()
		k = p.kids[a[i]]
		p.m.Unlock()
		if k == nil {
			return p, i
		}
		p = k
	}
	return p, len(a)
}

func (t *IPTrie) getData(byteSize int) *IPTrie {
	t.m.Lock()
	defer t.m.Unlock()
	var keys []int
	if byteSize >= 0 {
		for key := range t.kids {
			if int(key) < byteSize {
				keys = append(keys, int(key))
			}
		}
	} else {
		for key := range t.kids {
			keys = append(keys, int(key))
		}
	}
	sort.Ints(keys)
	var k *IPTrie
	var r *IPTrie
	for i := len(keys) - 1; i >= 0; i-- {
		k = t.kids[byte(keys[i])]
		r = k.getData(-1)
		if r == nil {
			continue
		}
		if r.data != nil || r.rangeStart != nil {
			return r
		}
	}
	if t.data != nil || t.rangeStart != nil {
		return t
	}
	return nil
}

// CIDRToRange computes the first and last IP address for a range from a valid
// CIDR string.  Because of the corner cases around a IPv4 prefix of length > 30
// (and to keep this function fast) any prefix of length > 30 will return
// nil for the start and end addresses.
func cidrToRange(addr string) (net.IP, net.IP) {
	ip, n, err := net.ParseCIDR(addr)
	if err != nil {
		return nil, nil
	}
	if pl, _ := n.Mask.Size(); pl > 30 {
		return nil, nil
	}
	al := len(ip)
	s := make(net.IP, al)
	e := make(net.IP, al)
	copy(s, ip)
	s[al-1] = s[al-1] | 0x01
	copy(e, s)
	d := al - len(n.Mask)
	for i := range n.Mask {
		e[i+d] = e[i+d] | (n.Mask[i] ^ 0xff)
	}
	e[al-1] -= 1
	return s, e
}

// IPv4ToUInt32 converts an IPv4 address to an uint32.
func IPv4ToUInt32(ip net.IP) uint32 {
	a := ip.To16()
	i := uint32(a[12]) << uint(24)
	i += uint32(a[13]) << uint(16)
	i += uint32(a[14]) << uint(8)
	i += uint32(a[15])
	return i
}

// Uint32ToIPv4 converts an uint32 to  an IPv4 address.
func Uint32ToIPv4(n uint32) net.IP {
	t := n
	a := make(net.IP, 16)
	a[15] = byte(t & 0x000000FF)
	t = t >> 8
	a[14] = byte(t & 0x000000FF)
	t = t >> 8
	a[13] = byte(t & 0x000000FF)
	t = t >> 8
	a[12] = byte(t & 0x000000FF)
	a[11] = 0xFF
	a[10] = 0xFF
	return a
}
