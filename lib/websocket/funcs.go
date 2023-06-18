// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"math/rand"
	"syscall"

	"golang.org/x/sys/unix"
)

// maxBuffer define maximum payload that we read/write from socket at one
// time.
// This number should be lower than MTU for better handling larger payload.
const maxBuffer = 1024

// Recv read all content from file descriptor into slice of bytes.
//
// On success it will return buffer from pool. Caller must put the buffer back
// to the pool.
//
// On fail it will return nil buffer and error.
func Recv(fd int) (packet []byte, err error) {
	var (
		buf []byte = make([]byte, maxBuffer)

		errno syscall.Errno
		n     int
		ok    bool
	)

	for {
		n, err = unix.Read(fd, buf)
		if err != nil {
			errno, ok = err.(unix.Errno)
			if ok {
				if errno == unix.EAGAIN || errno == unix.EINTR {
					continue
				}
			}
			break
		}
		if n > 0 {
			packet = append(packet, buf[:n]...)
		}
		if n < maxBuffer {
			break
		}
	}

	return packet, err
}

// Send the packet through web socket file descriptor `fd`.
func Send(fd int, packet []byte) (err error) {
	var (
		errno syscall.Errno
		max   int
		n     int
		ok    bool
	)

	for len(packet) > 0 {
		if len(packet) < maxBuffer {
			max = len(packet)
		} else {
			max = maxBuffer
		}

		n, err = unix.Write(fd, packet[:max])
		if err != nil {
			errno, ok = err.(unix.Errno)
			if ok {
				if errno == unix.EAGAIN {
					continue
				}
			}
			return err
		}

		if n > 0 {
			packet = packet[n:]
		}
	}

	return err
}

// generateHandshakeAccept generate server accept key by concatenating key,
// defined in step 4 in Section 4.2.2, with the string
// "258EAFA5-E914-47DA-95CA-C5AB0DC85B11", taking the SHA-1 hash of this
// concatenated value to obtain a 20-byte value and base64-encoding (see
// Section 4 of [RFC4648]) this 20-byte hash.
func generateHandshakeAccept(key []byte) string {
	key = append(key, "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"...)
	var sum [20]byte = sha1.Sum(key)
	return base64.StdEncoding.EncodeToString(sum[:])
}

// generateHandshakeKey randomly selected 16-byte value that has been
// base64-encoded (see Section 4 of [RFC4648]).
func generateHandshakeKey() (key []byte) {
	var bkey []byte = make([]byte, 16)

	binary.LittleEndian.PutUint64(bkey[0:8], rand.Uint64())
	binary.LittleEndian.PutUint64(bkey[8:16], rand.Uint64())

	key = make([]byte, base64.StdEncoding.EncodedLen(len(bkey)))
	base64.StdEncoding.Encode(key, bkey)

	return
}
