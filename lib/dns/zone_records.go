// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import "github.com/shuLhan/share/lib/reflect"

// zoneRecords contains mapping between domain name and its resource
// records.
type zoneRecords map[string][]*ResourceRecord

// add a ResourceRecord into the zone.
func (zr zoneRecords) add(rr *ResourceRecord) {
	listRR := zr[rr.Name]

	// Replace the RR if its type is SOA because only one SOA
	// should exist per domain name.
	if rr.Type == RecordTypeSOA {
		for x, in := range listRR {
			if in.Type != RecordTypeSOA {
				continue
			}
			listRR[x] = rr
			return
		}
	}
	listRR = append(listRR, rr)
	zr[rr.Name] = listRR
}

// remove a ResourceRecord from list by its Name and Value.
// It will return true if the RR exist and removed.
func (zr zoneRecords) remove(rr *ResourceRecord) bool {
	listRR := zr[rr.Name]
	for x, in := range listRR {
		if in.Type != rr.Type {
			continue
		}
		if in.Class != rr.Class {
			continue
		}
		if !reflect.IsEqual(in.Value, rr.Value) {
			continue
		}
		copy(listRR[x:], listRR[x+1:])
		listRR[len(listRR)-1] = nil
		zr[rr.Name] = listRR[:len(listRR)-1]
		return true
	}
	return false
}
