// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"

	libnet "git.sr.ht/~shulhan/pakakeh.go/lib/net"
)

// RDataSRV define a resource record for type SRV.
type RDataSRV struct {
	// The symbolic name of the desired service, as defined in Assigned
	// Numbers [STD 2] or locally.  An underscore (_) is prepended to
	// the service identifier to avoid collisions with DNS labels that
	// occur in nature.
	//
	// Some widely used services, notably POP, don't have a single
	// universal name.  If Assigned Numbers names the service
	// indicated, that name is the only name which is legal for SRV
	// lookups.  The Service is case insensitive.
	Service string

	// The symbolic name of the desired protocol, with an underscore
	// (_) prepended to prevent collisions with DNS labels that occur
	// in nature.  _TCP and _UDP are at present the most useful values
	// for this field, though any name defined by Assigned Numbers or
	// locally may be used (as for Service).  The Proto is case
	// insensitive.
	Proto string

	// The domain this RR refers to.  The SRV RR is unique in that the
	// name one searches for is not this name; the example near the end
	// shows this clearly.
	Name string

	// The domain name of the target host.  There MUST be one or more
	// address records for this name, the name MUST NOT be an alias (in
	// the sense of RFC 1034 or RFC 2181).  Implementors are urged, but
	// not required, to return the address record(s) in the Additional
	// Data section.  Unless and until permitted by future standards
	// action, name compression is not to be used for this field.
	//
	// A Target of "." means that the service is decidedly not
	// available at this domain.
	Target string

	// The priority of this target host.  A client MUST attempt to
	// contact the target host with the lowest-numbered priority it can
	// reach; target hosts with the same priority SHOULD be tried in an
	// order defined by the weight field.  The range is 0-65535.  This
	// is a 16 bit unsigned integer in network byte order.
	Priority uint16

	// A server selection mechanism.  The weight field specifies a
	// relative weight for entries with the same priority. Larger
	// weights SHOULD be given a proportionately higher probability of
	// being selected. The range of this number is 0-65535.  This is a
	// 16 bit unsigned integer in network byte order.  Domain
	// administrators SHOULD use Weight 0 when there isn't any server
	// selection to do, to make the RR easier to read for humans (less
	// noisy).  In the presence of records containing weights greater
	// than 0, records with weight 0 should have a very small chance of
	// being selected.
	//
	// In the absence of a protocol whose specification calls for the
	// use of other weighting information, a client arranges the SRV
	// RRs of the same Priority in the order in which target hosts,
	// specified by the SRV RRs, will be contacted. The following
	// algorithm SHOULD be used to order the SRV RRs of the same
	// priority:
	//
	// To select a target to be contacted next, arrange all SRV RRs
	// (that have not been ordered yet) in any order, except that all
	// those with weight 0 are placed at the beginning of the list.
	//
	// Compute the sum of the weights of those RRs, and with each RR
	// associate the running sum in the selected order. Then choose a
	// uniform random number between 0 and the sum computed
	// (inclusive), and select the RR whose running sum value is the
	// first in the selected order which is greater than or equal to
	// the random number selected. The target host specified in the
	// selected SRV RR is the next one to be contacted by the client.
	// Remove this SRV RR from the set of the unordered SRV RRs and
	// apply the described algorithm to the unordered SRV RRs to select
	// the next target host.  Continue the ordering process until there
	// are no unordered SRV RRs.  This process is repeated for each
	// Priority.
	Weight uint16

	// The port on this target host of this service.  The range is 0-
	// 65535.  This is a 16 bit unsigned integer in network byte order.
	// This is often as specified in Assigned Numbers but need not be.
	Port uint16
}

// String return readable representation of SRV record.
func (srv *RDataSRV) String() string {
	return fmt.Sprintf("{Service:%s Proto:%s Name:%s Priority:%d Weight:%d Port:%d Target:%s}",
		srv.Service, srv.Proto, srv.Name, srv.Priority, srv.Weight,
		srv.Port, srv.Target)
}

// initAndValidate initialize and validate the SRV fields.
func (srv *RDataSRV) initAndValidate() error {
	if !libnet.IsHostnameValid([]byte(srv.Target), true) {
		return fmt.Errorf("invalid SRV target %q", srv.Target)
	}
	return nil
}
