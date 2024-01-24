// Copyright 2024, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path_test

import (
	"fmt"
	"log"

	libpath "github.com/shuLhan/share/lib/path"
)

func ExampleRoute_Parse() {
	var (
		rute *libpath.Route
		err  error
	)

	rute, err = libpath.NewRoute(`/book/:title/:page`)
	if err != nil {
		log.Fatal(err)
	}

	var (
		vals map[string]string
		ok   bool
	)

	vals, ok = rute.Parse(`/book/Hitchiker to Galaxy/42`)
	fmt.Println(ok, vals)

	vals, ok = rute.Parse(`/book/Hitchiker to Galaxy`)
	fmt.Println(ok, vals)

	vals, ok = rute.Parse(`/book/Hitchiker to Galaxy/42/order`)
	fmt.Println(ok, vals)

	// Output:
	// true map[page:42 title:hitchiker to galaxy]
	// false map[]
	// false map[]
}

func ExampleRoute_Set() {
	var (
		rute *libpath.Route
		err  error
	)
	rute, err = libpath.NewRoute(`/:user/:repo`)
	if err != nil {
		log.Fatal(err)
	}

	rute.Set(`user`, `shuLhan`)
	fmt.Println(rute)

	rute.Set(`repo`, `share`)
	fmt.Println(rute)
	// Output:
	// /shuLhan/:repo
	// /shuLhan/share
}
