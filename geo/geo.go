// Copyright 2013 The iptrie Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The geo package provides functions for loading maxmind geolocation data into
// an IPTrie so that lookups can be handled quickly.  For more information on
// maxmind data sources please see http://www.maxmind.com/en/opensource.
package geo

import (
	"encoding/csv"
	"io"
	"strconv"

	"code.google.com/p/iptrie"
)

// A Loc represents a location.  Not all fields will be filled when there is
// incomplete data for that location.
type Loc struct {
	CountryCode string // ISO 3166-1 alpha-2 code
	Region      string
	City        string
	Lat         float64
	Lon         float64
}

// An AS represnets the number (ASN) and description for an autonomous system.
type AS struct {
	Num int64  // AS number
	Dsc string // AS description
}

// AddMaxmindIPv6ASN reads CSV information from the maxmind IPv6 ASN block
// file and adds the appropriate ranges to the IPTrie.
func AddMaxmindIPv6ASN(t *iptrie.IPTrie, block io.Reader) error {
	c := csv.NewReader(block)
	c.FieldsPerRecord = -1
	c.TrailingComma = true
	var a *AS
	var r []string
	var err error
	for {
		r, err = c.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if len(r) < 6 {
			continue
		}
		a = &AS{
			Dsc: r[5],
		}
		a.Num, _ = strconv.ParseInt(r[4], 10, 64)
		t.AddRange(r[0], r[1], a)
	}
	return nil
}

// AddMaxmindIPv4ASN reads CSV information from the maxmind IPv4 ASN block
// file and adds the appropriate ranges to the IPTrie.
func AddMaxmindIPv4ASN(t *iptrie.IPTrie, block io.Reader) error {
	c := csv.NewReader(block)
	c.FieldsPerRecord = -1
	c.TrailingComma = true
	var a *AS
	var r []string
	var s int64
	var e int64
	var err error
	for {
		r, err = c.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if len(r) < 4 {
			continue
		}
		a = &AS{
			Dsc: r[3],
		}
		a.Num, _ = strconv.ParseInt(r[2], 10, 64)
		s, err = strconv.ParseInt(r[0], 10, 64)
		if err != nil {
			continue
		}
		e, err = strconv.ParseInt(r[1], 10, 64)
		if err != nil {
			continue
		}
		t.AddRangeNum(uint32(s), uint32(e), a)
	}
	return nil
}

// AddMaxmindIPv6City reads CSV information from the maxmind IPv6 city block
// file and adds the appropriate ranges to the IPTrie.
func AddMaxmindIPv6City(t *iptrie.IPTrie, block io.Reader) error {
	c := csv.NewReader(block)
	c.FieldsPerRecord = -1
	c.TrailingComma = true
	var loc *Loc
	var r []string
	var err error
	for {
		r, err = c.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if len(r) < 9 {
			continue
		}
		loc = &Loc{
			CountryCode: r[4],
			Region:      r[5],
		}
		loc.Lat, _ = strconv.ParseFloat(r[7], 64)
		loc.Lon, _ = strconv.ParseFloat(r[8], 64)
		t.AddRange(r[0], r[1], loc)
	}
	return nil
}

// AddMaxmindIPv4City reads CSV information from the maxmind IPv6 city block
// and location file and adds the appropriate ranges to the IPTrie.
func AddMaxmindIPv4City(t *iptrie.IPTrie, block, location io.Reader) error {
	c := csv.NewReader(location)
	c.FieldsPerRecord = -1
	c.TrailingComma = true
	lm := make(map[string]*Loc)
	var r []string
	var err error
	for {
		r, err = c.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if len(r) < 7 {
			continue
		}
		lm[r[0]] = &Loc{
			CountryCode: r[1],
			Region:      r[2],
			City:        r[3],
		}

		lm[r[0]].Lat, _ = strconv.ParseFloat(r[5], 64)
		lm[r[0]].Lon, _ = strconv.ParseFloat(r[6], 64)
	}
	c = csv.NewReader(block)
	c.FieldsPerRecord = -1
	c.TrailingComma = true
	var loc *Loc
	var s int64
	var e int64
	for {
		r, err = c.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if len(r) < 3 {
			continue
		}
		loc = lm[r[2]]
		if loc == nil {
			continue
		}
		s, err = strconv.ParseInt(r[0], 10, 64)
		if err != nil {
			continue
		}
		e, err = strconv.ParseInt(r[1], 10, 64)
		if err != nil {
			continue
		}
		t.AddRangeNum(uint32(s), uint32(e), loc)
	}
	return nil
}

// AddMaxmindIPv4City reads CSV information from the maxmind IPv6 country
// block and location file and adds the appropriate ranges to the IPTrie.
// Don't add this data with the city data, but instead use it in a separate
// iptrie.IPTrie to fill in nil results.
func AddMaxmindIPv4Country(t *iptrie.IPTrie, block, location io.Reader) error {
	c := csv.NewReader(location)
	c.FieldsPerRecord = -1
	c.TrailingComma = true
	lm := make(map[string]*Loc)
	var r []string
	var err error
	for {
		r, err = c.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if len(r) < 4 {
			continue
		}
		lm[r[0]] = &Loc{
			CountryCode: r[0],
		}

		lm[r[0]].Lat, _ = strconv.ParseFloat(r[1], 64)
		lm[r[0]].Lon, _ = strconv.ParseFloat(r[2], 64)
	}
	c = csv.NewReader(block)
	c.FieldsPerRecord = -1
	c.TrailingComma = true
	var loc *Loc
	for {
		r, err = c.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if len(r) < 5 {
			continue
		}
		loc = lm[r[4]]
		if loc == nil {
			continue
		}
		t.AddRange(r[0], r[1], loc)
	}
	return nil
}
