// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Program mimeunquote convert the email body from quoted-printable to plain
// text.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/quotedprintable"
	"net/mail"
	"os"
)

func main() {
	flag.Parse()

	var fileInput = flag.Arg(0)
	if len(fileInput) == 0 {
		log.Fatalf(`missing file input`)
	}

	var (
		fin *os.File
		err error
	)

	fin, err = os.Open(fileInput)
	if err != nil {
		log.Fatal(err)
	}

	var msg *mail.Message

	msg, err = mail.ReadMessage(fin)
	if err != nil {
		log.Fatal(err)
	}

	var (
		buf = make([]byte, 1024)
		qpr = quotedprintable.NewReader(msg.Body)

		n int
	)

	for {
		n, err = qpr.Read(buf)
		if n != 0 {
			fmt.Printf(`%s`, buf[:n])
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(err)
		}
		if n == 0 {
			break
		}
	}
}
