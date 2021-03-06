// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestZoneParseDirectiveOrigin(t *testing.T) {
	cases := []struct {
		desc   string
		in     string
		expErr string
		exp    string
	}{{
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

	m := newZoneParser("")

	for _, c := range cases {
		t.Log(c.desc)

		m.Init(c.in, "", 0)

		err := m.parse()
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "origin", c.exp, m.origin, true)
	}
}

func TestZoneParseDirectiveInclude(t *testing.T) {
	cases := []struct {
		desc   string
		in     string
		expErr string
	}{{
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

	m := newZoneParser("")

	for _, c := range cases {
		t.Log(c.desc)

		m.Init(c.in, "", 0)

		err := m.parse()
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
			continue
		}
	}
}

func TestZoneParseDirectiveTTL(t *testing.T) {
	cases := []struct {
		desc   string
		in     string
		expErr string
		exp    uint32
	}{{
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

	m := newZoneParser("")

	for _, c := range cases {
		t.Log(c.desc)

		m.Init(c.in, "", 0)

		err := m.parse()
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "ttl", c.exp, m.ttl, true)
	}
}

func TestZoneInitRFC1035(t *testing.T) {
	cases := []struct {
		desc   string
		origin string
		ttl    uint32
		in     string
		expErr error
		exp    []*Message
	}{{
		desc:   "RFC1035 section 5.3",
		origin: "ISI.EDU",
		ttl:    3600,
		in: `
@   IN  SOA     VENERA      Action\.domains (
                                 20     ; SERIAL
                                 7200   ; REFRESH
                                 600    ; RETRY
                                 3600000; EXPIRE
                                 60)    ; MINIMUM

        NS      A.ISI.EDU.
        NS      VENERA
        NS      VAXA
        MX      10      VENERA
        MX      20      VAXA

A       A       26.3.0.103

VENERA  A       10.1.0.52
        A       128.9.0.32

VAXA    A       10.2.0.27
        A       128.9.0.33

`,
		exp: []*Message{{
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "isi.edu",
				Type:  QueryTypeSOA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "isi.edu",
				Type:  QueryTypeSOA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: &RDataSOA{
					MName:   "venera.isi.edu",
					RName:   "action\\.domains.isi.edu",
					Serial:  20,
					Refresh: 7200,
					Retry:   600,
					Expire:  3600000,
					Minimum: 60,
				},
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 3,
			},
			Question: SectionQuestion{
				Name:  "isi.edu",
				Type:  QueryTypeNS,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "isi.edu",
				Type:  QueryTypeNS,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "a.isi.edu",
			}, {
				Name:  "isi.edu",
				Type:  QueryTypeNS,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "venera.isi.edu",
			}, {
				Name:  "isi.edu",
				Type:  QueryTypeNS,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "vaxa.isi.edu",
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 2,
			},
			Question: SectionQuestion{
				Name:  "isi.edu",
				Type:  QueryTypeMX,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "isi.edu",
				Type:  QueryTypeMX,
				Class: QueryClassIN,
				TTL:   3600,
				Value: &RDataMX{
					Preference: 10,
					Exchange:   "venera.isi.edu",
				},
			}, {
				Name:  "isi.edu",
				Type:  QueryTypeMX,
				Class: QueryClassIN,
				TTL:   3600,
				Value: &RDataMX{
					Preference: 20,
					Exchange:   "vaxa.isi.edu",
				},
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "a.isi.edu",
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "a.isi.edu",
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "26.3.0.103",
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 2,
			},
			Question: SectionQuestion{
				Name:  "venera.isi.edu",
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "venera.isi.edu",
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "10.1.0.52",
			}, {
				Name:  "venera.isi.edu",
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "128.9.0.32",
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 2,
			},
			Question: SectionQuestion{
				Name:  "vaxa.isi.edu",
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "vaxa.isi.edu",
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "10.2.0.27",
			}, {
				Name:  "vaxa.isi.edu",
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "128.9.0.33",
			}},
		}},
	}}

	m := newZoneParser("")

	for _, c := range cases {
		t.Log(c.desc)

		m.Init(c.in, c.origin, c.ttl)

		err := m.parse()
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "messages length:",
			len(c.exp), len(m.zone.messages), true)

		for x, msg := range m.zone.messages {
			test.Assert(t, "Message.Header",
				c.exp[x].Header, msg.Header, true)
			test.Assert(t, "Message.Question",
				c.exp[x].Question, msg.Question, true)

			for y, answer := range msg.Answer {
				test.Assert(t, "Answer.Name",
					c.exp[x].Answer[y].Name,
					answer.Name, true)
				test.Assert(t, "Answer.Type",
					c.exp[x].Answer[y].Type,
					answer.Type, true)
				test.Assert(t, "Answer.Class",
					c.exp[x].Answer[y].Class,
					answer.Class, true)
				test.Assert(t, "Answer.TTL",
					c.exp[x].Answer[y].TTL,
					answer.TTL, true)
				test.Assert(t, "Answer.Value",
					c.exp[x].Answer[y].Value,
					answer.Value, true)
			}
			for y, auth := range msg.Authority {
				test.Assert(t, "Authority.Name",
					c.exp[x].Authority[y].Name,
					auth.Name, true)
				test.Assert(t, "Authority.Type",
					c.exp[x].Authority[y].Type,
					auth.Type, true)
				test.Assert(t, "Authority.Class",
					c.exp[x].Authority[y].Class,
					auth.Class, true)
				test.Assert(t, "Authority.TTL",
					c.exp[x].Authority[y].TTL,
					auth.TTL, true)
				test.Assert(t, "Authority.Value",
					c.exp[x].Authority[y].Value, auth.Value, true)
			}
			for y, add := range msg.Additional {
				test.Assert(t, "Additional.Name",
					c.exp[x].Additional[y].Name,
					add.Name, true)
				test.Assert(t, "Additional.Type",
					c.exp[x].Additional[y].Type,
					add.Type, true)
				test.Assert(t, "Additional.Class",
					c.exp[x].Additional[y].Class,
					add.Class, true)
				test.Assert(t, "Additional.TTL",
					c.exp[x].Additional[y].TTL,
					add.TTL, true)
				test.Assert(t, "Additional.Value",
					c.exp[x].Additional[y].Value,
					add.Value, true)
			}
		}
	}
}

func TestZoneInit2(t *testing.T) {
	cases := []struct {
		desc   string
		origin string
		ttl    uint32
		in     string
		expErr error
		exp    []*Message
	}{{
		desc: "From http://www.tcpipguide.com/free/t_DNSZoneFileFormat-4.htm",
		in: `
$ORIGIN pcguide.com.
@ IN SOA ns23.pair.com. root.pair.com. (
2001072300 ; Serial
3600 ; Refresh
300 ; Retry
604800 ; Expire
3600 ) ; Minimum

@ IN NS ns23.pair.com.
@ IN NS ns0.ns0.com.

localhost IN A 127.0.0.1
@ IN A 209.68.14.80
  IN MX 50 qs939.pair.com.

www IN CNAME @
ftp IN CNAME @
mail IN CNAME @
relay IN CNAME relay.pair.com.
`,
		exp: []*Message{{
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "pcguide.com",
				Type:  QueryTypeSOA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "pcguide.com",
				Type:  QueryTypeSOA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: &RDataSOA{
					MName:   "ns23.pair.com",
					RName:   "root.pair.com",
					Serial:  2001072300,
					Refresh: 3600,
					Retry:   300,
					Expire:  604800,
					Minimum: 3600,
				},
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 2,
			},
			Question: SectionQuestion{
				Name:  "pcguide.com",
				Type:  QueryTypeNS,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "pcguide.com",
				Type:  QueryTypeNS,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "ns23.pair.com",
			}, {
				Name:  "pcguide.com",
				Type:  QueryTypeNS,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "ns0.ns0.com",
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "localhost.pcguide.com",
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "localhost.pcguide.com",
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "127.0.0.1",
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "pcguide.com",
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "pcguide.com",
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "209.68.14.80",
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "pcguide.com",
				Type:  QueryTypeMX,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "pcguide.com",
				Type:  QueryTypeMX,
				Class: QueryClassIN,
				TTL:   3600,
				Value: &RDataMX{
					Preference: 50,
					Exchange:   "qs939.pair.com",
				},
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "www.pcguide.com",
				Type:  QueryTypeCNAME,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "www.pcguide.com",
				Type:  QueryTypeCNAME,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "pcguide.com",
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "ftp.pcguide.com",
				Type:  QueryTypeCNAME,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "ftp.pcguide.com",
				Type:  QueryTypeCNAME,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "pcguide.com",
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "mail.pcguide.com",
				Type:  QueryTypeCNAME,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "mail.pcguide.com",
				Type:  QueryTypeCNAME,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "pcguide.com",
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "relay.pcguide.com",
				Type:  QueryTypeCNAME,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "relay.pcguide.com",
				Type:  QueryTypeCNAME,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "relay.pair.com",
			}},
		}},
	}}

	m := newZoneParser("")

	for _, c := range cases {
		t.Log(c.desc)

		m.Init(c.in, c.origin, c.ttl)

		err := m.parse()
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "messages length:", len(c.exp),
			len(m.zone.messages), true)

		for x, msg := range m.zone.messages {
			test.Assert(t, "Message.Header",
				c.exp[x].Header, msg.Header, true)
			test.Assert(t, "Message.Question",
				c.exp[x].Question, msg.Question, true)

			for y, answer := range msg.Answer {
				test.Assert(t, "Answer.Name",
					c.exp[x].Answer[y].Name, answer.Name,
					true)
				test.Assert(t, "Answer.Type",
					c.exp[x].Answer[y].Type, answer.Type,
					true)
				test.Assert(t, "Answer.Class",
					c.exp[x].Answer[y].Class,
					answer.Class, true)
				test.Assert(t, "Answer.TTL",
					c.exp[x].Answer[y].TTL,
					answer.TTL, true)
				test.Assert(t, "Answer.Value",
					c.exp[x].Answer[y].Value,
					answer.Value, true)
			}
			for y, auth := range msg.Authority {
				test.Assert(t, "Authority.Name",
					c.exp[x].Authority[y].Name,
					auth.Name, true)
				test.Assert(t, "Authority.Type",
					c.exp[x].Authority[y].Type,
					auth.Type, true)
				test.Assert(t, "Authority.Class",
					c.exp[x].Authority[y].Class,
					auth.Class, true)
				test.Assert(t, "Authority.TTL",
					c.exp[x].Authority[y].TTL,
					auth.TTL, true)
				test.Assert(t, "Authority.Value",
					c.exp[x].Authority[y].Value,
					auth.Value, true)
			}
			for y, add := range msg.Additional {
				test.Assert(t, "Additional.Name",
					c.exp[x].Additional[y].Name,
					add.Name, true)
				test.Assert(t, "Additional.Type",
					c.exp[x].Additional[y].Type,
					add.Type, true)
				test.Assert(t, "Additional.Class",
					c.exp[x].Additional[y].Class,
					add.Class, true)
				test.Assert(t, "Additional.TTL",
					c.exp[x].Additional[y].TTL,
					add.TTL, true)
				test.Assert(t, "Additional.Value",
					c.exp[x].Additional[y].Value,
					add.Value, true)
			}
		}
	}
}

func TestZoneInit3(t *testing.T) {
	cases := []struct {
		desc   string
		origin string
		ttl    uint32
		in     string
		expErr error
		exp    []*Message
	}{{
		desc:   "From http://www.tcpipguide.com/free/t_DNSZoneFileFormat-4.htm",
		origin: "localdomain",
		in: `
; Applications.
dev.kilabit.info.  A  127.0.0.1
dev.kilabit.com.   A  127.0.0.1

; Documentations.
angularjs.doc       A  127.0.0.1
`,
		exp: []*Message{{
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "dev.kilabit.info",
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "dev.kilabit.info",
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "127.0.0.1",
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "dev.kilabit.com",
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "dev.kilabit.com",
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "127.0.0.1",
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "angularjs.doc.localdomain",
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "angularjs.doc.localdomain",
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "127.0.0.1",
			}},
		}},
	}}

	m := newZoneParser("")

	for _, c := range cases {
		t.Log(c.desc)

		m.Init(c.in, c.origin, c.ttl)

		err := m.parse()
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "messages length:", len(c.exp),
			len(m.zone.messages), true)

		for x, msg := range m.zone.messages {
			test.Assert(t, "Message.Header", c.exp[x].Header,
				msg.Header, true)
			test.Assert(t, "Message.Question",
				c.exp[x].Question, msg.Question, true)

			for y, answer := range msg.Answer {
				test.Assert(t, "Answer.Name",
					c.exp[x].Answer[y].Name, answer.Name,
					true)
				test.Assert(t, "Answer.Type",
					c.exp[x].Answer[y].Type, answer.Type,
					true)
				test.Assert(t, "Answer.Class",
					c.exp[x].Answer[y].Class,
					answer.Class, true)
				test.Assert(t, "Answer.TTL",
					c.exp[x].Answer[y].TTL, answer.TTL,
					true)
				test.Assert(t, "Answer.Value",
					c.exp[x].Answer[y].Value,
					answer.Value, true)
			}
			for y, auth := range msg.Authority {
				test.Assert(t, "Authority.Name",
					c.exp[x].Authority[y].Name,
					auth.Name, true)
				test.Assert(t, "Authority.Type",
					c.exp[x].Authority[y].Type,
					auth.Type, true)
				test.Assert(t, "Authority.Class",
					c.exp[x].Authority[y].Class,
					auth.Class, true)
				test.Assert(t, "Authority.TTL",
					c.exp[x].Authority[y].TTL, auth.TTL,
					true)
				test.Assert(t, "Authority.Value",
					c.exp[x].Authority[y].Value,
					auth.Value, true)
			}
			for y, add := range msg.Additional {
				test.Assert(t, "Additional.Name",
					c.exp[x].Additional[y].Name,
					add.Name, true)
				test.Assert(t, "Additional.Type",
					c.exp[x].Additional[y].Type,
					add.Type, true)
				test.Assert(t, "Additional.Class",
					c.exp[x].Additional[y].Class,
					add.Class, true)
				test.Assert(t, "Additional.TTL",
					c.exp[x].Additional[y].TTL, add.TTL,
					true)
				test.Assert(t, "Additional.Value",
					c.exp[x].Additional[y].Value,
					add.Value, true)
			}
		}
	}
}

func TestZoneParseTXT(t *testing.T) {
	cases := []struct {
		in       string
		exp      []*Message
		expError string
	}{{
		in: `@ IN TXT "This is a test"`,
		exp: []*Message{{
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "kilabit.local",
				Type:  QueryTypeTXT,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "kilabit.local",
				Type:  QueryTypeTXT,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "This is a test",
			}},
		}},
	}}

	m := newZoneParser("")

	for _, c := range cases {
		m.Init(c.in, "kilabit.local", 3600)

		err := m.parse()
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error(), true)
			continue
		}

		test.Assert(t, "messages length:", len(c.exp),
			len(m.zone.messages), true)

		for x, msg := range m.zone.messages {
			test.Assert(t, "Message.Header", c.exp[x].Header,
				msg.Header, true)
			test.Assert(t, "Message.Question", c.exp[x].Question,
				msg.Question, true)

			for y, answer := range msg.Answer {
				test.Assert(t, "Answer.Name",
					c.exp[x].Answer[y].Name, answer.Name,
					true)
				test.Assert(t, "Answer.Type",
					c.exp[x].Answer[y].Type, answer.Type,
					true)
				test.Assert(t, "Answer.Class",
					c.exp[x].Answer[y].Class,
					answer.Class, true)
				test.Assert(t, "Answer.TTL",
					c.exp[x].Answer[y].TTL, answer.TTL,
					true)
				test.Assert(t, "Answer.Value",
					c.exp[x].Answer[y].Value,
					answer.Value, true)
			}
			for y, auth := range msg.Authority {
				test.Assert(t, "Authority.Name",
					c.exp[x].Authority[y].Name, auth.Name,
					true)
				test.Assert(t, "Authority.Type",
					c.exp[x].Authority[y].Type, auth.Type,
					true)
				test.Assert(t, "Authority.Class",
					c.exp[x].Authority[y].Class,
					auth.Class, true)
				test.Assert(t, "Authority.TTL",
					c.exp[x].Authority[y].TTL, auth.TTL,
					true)
				test.Assert(t, "Authority.Value",
					c.exp[x].Authority[y].Value,
					auth.Value, true)
			}
			for y, add := range msg.Additional {
				test.Assert(t, "Additional.Name",
					c.exp[x].Additional[y].Name,
					add.Name, true)
				test.Assert(t, "Additional.Type",
					c.exp[x].Additional[y].Type,
					add.Type, true)
				test.Assert(t, "Additional.Class",
					c.exp[x].Additional[y].Class,
					add.Class, true)
				test.Assert(t, "Additional.TTL",
					c.exp[x].Additional[y].TTL,
					add.TTL, true)
				test.Assert(t, "Additional.Value",
					c.exp[x].Additional[y].Value,
					add.Value, true)
			}
		}
	}
}
