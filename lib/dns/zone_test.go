// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/test"
)

func TestParseZone(t *testing.T) {
	var (
		listTData []*test.Data
		err       error
	)

	listTData, err = test.LoadDataDir(`testdata/zone/`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		tdata  *test.Data
		zone   *Zone
		bb     bytes.Buffer
		origin string
		vstr   string
		vbytes []byte
		ttl    int64
	)
	for _, tdata = range listTData {
		t.Log(tdata.Name)

		origin = tdata.Flag[`origin`]

		vstr = tdata.Flag[`ttl`]
		if len(vstr) > 0 {
			ttl, err = strconv.ParseInt(vstr, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
		} else {
			ttl = 0
		}

		vbytes = tdata.Input[`zone_in.txt`]
		zone, err = ParseZone(vbytes, origin, uint32(ttl))
		if err != nil {
			t.Fatal(err)
		}

		// Compare the zone by writing back to text.

		bb.Reset()
		_, err = zone.WriteTo(&bb)
		if err != nil {
			t.Fatal(err)
		}

		vbytes = tdata.Output[`zone_out.txt`]
		test.Assert(t, tdata.Name, string(vbytes), bb.String())

		// Compare the packed zone as message.

		var (
			msg *Message
			tag string
			x   int
		)
		for x, msg = range zone.messages {
			bb.Reset()
			libbytes.DumpPrettyTable(&bb, msg.Question.String(), msg.packet)

			tag = fmt.Sprintf(`message_%d.hex`, x)
			vbytes = tdata.Output[tag]
			test.Assert(t, tag, string(vbytes), bb.String())
		}
	}
}

func TestZoneParseDirectiveOrigin(t *testing.T) {
	type testCase struct {
		desc   string
		in     string
		expErr string
		exp    string
	}

	var (
		m = newZoneParser(nil)

		cases []testCase
		c     testCase
		err   error
	)

	cases = []testCase{{
		desc:   "Without value",
		in:     `$origin`,
		expErr: "line 1: empty $origin directive",
	}, {
		desc:   "Without value and with comment",
		in:     `$origin ; comment`,
		expErr: "line 1: empty $origin directive",
	}, {
		desc: "With value",
		in:   `$origin x`,
		exp:  "x",
	}, {
		desc: "With value and comment",
		in:   `$origin x ;comment`,
		exp:  "x",
	}}

	for _, c = range cases {
		t.Log(c.desc)

		m.Reset([]byte(c.in), ``, 0)

		err = m.parse()
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "origin", c.exp, m.origin)
	}
}

func TestZoneParseDirectiveInclude(t *testing.T) {
	type testCase struct {
		desc   string
		in     string
		expErr string
	}

	var (
		m = newZoneParser(nil)

		cases []testCase
		c     testCase
		err   error
	)

	cases = []testCase{{
		desc:   "Without value",
		in:     `$include`,
		expErr: "line 1: empty $include directive",
	}, {
		desc:   "Without value and with comment",
		in:     `$include ; comment`,
		expErr: "line 1: empty $include directive",
	}, {
		desc: "With value",
		in:   `$include testdata/sub.domain`,
	}, {
		desc: "With value and comment",
		in:   `$origin testdata/sub.domain ;comment`,
	}}

	for _, c = range cases {
		t.Log(c.desc)

		m.Reset([]byte(c.in), ``, 0)

		err = m.parse()
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error())
			continue
		}
	}
}

func TestZoneParseDirectiveTTL(t *testing.T) {
	type testCase struct {
		desc   string
		in     string
		expErr string
		exp    uint32
	}

	var (
		m = newZoneParser(nil)

		cases []testCase
		c     testCase
		err   error
	)

	cases = []testCase{{
		desc:   "Without value",
		in:     `$ttl`,
		expErr: "line 1: empty $TTL directive",
	}, {
		desc:   "Without value and with comment",
		in:     `$ttl ; comment`,
		expErr: "line 1: empty $TTL directive",
	}, {
		desc: "With value",
		in:   `$ttl 1`,
		exp:  1,
	}, {
		desc: "With value and comment",
		in:   `$ttl 1 ;comment`,
		exp:  1,
	}}

	for _, c = range cases {
		t.Log(c.desc)

		m.Reset([]byte(c.in), ``, 0)

		err = m.parse()
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "ttl", c.exp, m.ttl)
	}
}

func TestZoneParseTXT(t *testing.T) {
	type testCase struct {
		in       string
		expError string
		exp      []*Message
	}

	var (
		m = newZoneParser(nil)

		msg   *Message
		rr    ResourceRecord
		cases []testCase
		c     testCase
		err   error
		x, y  int
	)

	cases = []testCase{{
		in: `@ IN TXT "This is a test"`,
		exp: []*Message{{
			Header: MessageHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: MessageQuestion{
				Name:  `kilabit.local.`,
				Type:  RecordTypeTXT,
				Class: RecordClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  `kilabit.local.`,
				Type:  RecordTypeTXT,
				Class: RecordClassIN,
				TTL:   3600,
				Value: "This is a test",
			}},
		}},
	}}

	for _, c = range cases {
		m.Reset([]byte(c.in), `kilabit.local`, 3600)

		err = m.parse()
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error())
			continue
		}

		test.Assert(t, "messages length:", len(c.exp), len(m.zone.messages))

		for x, msg = range m.zone.messages {
			test.Assert(t, "Message.Header", c.exp[x].Header, msg.Header)
			test.Assert(t, "Message.Question", c.exp[x].Question, msg.Question)

			for y, rr = range msg.Answer {
				test.Assert(t, "Answer.Name", c.exp[x].Answer[y].Name, rr.Name)
				test.Assert(t, "Answer.Type", c.exp[x].Answer[y].Type, rr.Type)
				test.Assert(t, "Answer.Class", c.exp[x].Answer[y].Class, rr.Class)
				test.Assert(t, "Answer.TTL", c.exp[x].Answer[y].TTL, rr.TTL)
				test.Assert(t, "Answer.Value", c.exp[x].Answer[y].Value, rr.Value)
			}
			for y, rr = range msg.Authority {
				test.Assert(t, "Authority.Name", c.exp[x].Authority[y].Name, rr.Name)
				test.Assert(t, "Authority.Type", c.exp[x].Authority[y].Type, rr.Type)
				test.Assert(t, "Authority.Class", c.exp[x].Authority[y].Class, rr.Class)
				test.Assert(t, "Authority.TTL", c.exp[x].Authority[y].TTL, rr.TTL)
				test.Assert(t, "Authority.Value", c.exp[x].Authority[y].Value, rr.Value)
			}
			for y, rr = range msg.Additional {
				test.Assert(t, "Additional.Name", c.exp[x].Additional[y].Name, rr.Name)
				test.Assert(t, "Additional.Type", c.exp[x].Additional[y].Type, rr.Type)
				test.Assert(t, "Additional.Class", c.exp[x].Additional[y].Class, rr.Class)
				test.Assert(t, "Additional.TTL", c.exp[x].Additional[y].TTL, rr.TTL)
				test.Assert(t, "Additional.Value", c.exp[x].Additional[y].Value, rr.Value)
			}
		}
	}
}
