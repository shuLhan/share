// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"errors"
	"log"
	"sync"
)

// UDPClientPool contains a pool of UDP client connections.
//
// Any implementation that access UDPClient in multiple Go routines should
// create one client per Go routine; instead of using a single UDP
// client.
// The reason for this is because UDP packet is asynchronous.
//
// WARNING: using pooling is only works if client only call Lookup or Query.
// If implementation call Send() n client connection, make sure, it also call
// Recv on the same routine before putting the client back to pool.
type UDPClientPool struct {
	pool *sync.Pool
	ns   []string
	seq  int // sequence for the next name server.
	sync.Mutex
}

// NewUDPClientPool create pool for UDP connection using list of name servers.
// If no name servers is defined it will return nil.
func NewUDPClientPool(nameServers []string) (ucp *UDPClientPool, err error) {
	if len(nameServers) == 0 {
		err = errors.New("udp: UDPClientPool: no name servers defined")
		return nil, err
	}
	ucp = &UDPClientPool{
		ns: nameServers,
	}
	ucp.pool = &sync.Pool{
		New: ucp.newClient,
	}

	var (
		cl *UDPClient
		x  int
	)

	// Create new client for each name server, and push it to pool.
	// This is required to check if each name server is a valid IP
	// address because we did not want the New method return nil client
	// later.
	for x = 0; x < len(nameServers); x++ {
		cl, err = NewUDPClient(nameServers[x])
		if err != nil {
			return nil, err
		}
		ucp.Put(cl)
	}

	return ucp, nil
}

// newClient create a new udp client.
func (ucp *UDPClientPool) newClient() interface{} {
	var (
		cl  *UDPClient
		err error
	)

	ucp.Lock()
	defer ucp.Unlock()

	ucp.seq %= len(ucp.ns)
	cl, err = NewUDPClient(ucp.ns[ucp.seq])
	if err != nil {
		log.Fatal("udp: UDPClientPool: cannot create new client: ", err)
	}
	ucp.seq++
	return cl
}

// Get return UDP client.
func (ucp *UDPClientPool) Get() *UDPClient {
	return ucp.pool.Get().(*UDPClient)
}

// Put the UDP client into pool.
//
// WARNING: any client connection that call Send(), MUST call Recv()
// before putting client back to pool.  You have been warned.
func (ucp *UDPClientPool) Put(cl interface{}) {
	if cl != nil {
		ucp.pool.Put(cl)
	}
}
