// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paseto

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"log"
)

func ExamplePublicMode() {
	senderSK, _ := hex.DecodeString("e9ae9c7eae2fce6fd6727b5ca8df0fbc0aa60a5ffb354d4fdee1729e4e1463688d2160a4dc71a9a697d6ad6424da3f9dd18a259cdd51b0ae2b521e998b82d36e")
	senderPK, _ := hex.DecodeString("8d2160a4dc71a9a697d6ad6424da3f9dd18a259cdd51b0ae2b521e998b82d36e")
	senderKey := Key{
		id:      "sender",
		private: ed25519.PrivateKey(senderSK),
		public:  ed25519.PublicKey(senderPK),
	}

	receiverSK, _ := hex.DecodeString("4983da648bff1fd3e1892df9c56370215aa640829a5cab02d6616b115fa0bc5707c22e74ab9b181f8d87bdf03cf88476ec4c35e5517e173f236592f6695d59f5")
	receiverPK, _ := hex.DecodeString("07c22e74ab9b181f8d87bdf03cf88476ec4c35e5517e173f236592f6695d59f5")
	receiverKey := Key{
		id:      "receiver",
		private: ed25519.PrivateKey(receiverSK),
		public:  ed25519.PublicKey(receiverPK),
	}

	//
	// In the sender part, we register the sender key and the public key
	// of receiver in the list of peers.
	//
	senderPeers := map[string]ed25519.PublicKey{
		receiverKey.id: receiverKey.public,
	}
	sender := NewPublicMode(senderKey, senderPeers)

	addFooter := map[string]interface{}{
		"FOOTER": "HERE",
	}
	token, err := sender.Pack([]byte("hello receiver"), addFooter)
	if err != nil {
		log.Fatal(err)
	}

	// token generated by sender and send to receiver
	// ...

	//
	// In the receiver part, we register the receiver key and the public key
	// of sender in the list of peers.
	//
	receiverPeers := map[string]ed25519.PublicKey{
		senderKey.id: senderKey.public,
	}
	receiver := NewPublicMode(receiverKey, receiverPeers)

	// receiver receive the token from sender and unpack it ...
	gotData, gotFooter, err := receiver.Unpack(token)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Received data: %s\n", gotData)
	fmt.Printf("Received footer: %+v\n", gotFooter)
	// Output:
	// Received data: hello receiver
	// Received footer: map[FOOTER:HERE]
}
