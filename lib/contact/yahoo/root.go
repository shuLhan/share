// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yahoo

// Root define the root of JSON response.
type Root struct {
	Contacts Contacts `json:"contacts"`
}
