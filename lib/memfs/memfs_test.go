// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
	"git.sr.ht/~shulhan/pakakeh.go/lib/text/diff"
)

var (
	_testWD string
)

func TestMain(m *testing.M) {
	var err error

	_testWD, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(filepath.Join(_testWD, "testdata/exclude/dir"), 0700)
	if err != nil {
		var perr *fs.PathError
		if !errors.As(err, &perr) {
			log.Fatal("!ok:", err)
		}
		if !errors.Is(perr.Err, os.ErrExist) {
			log.Fatalf("perr: %+v %+v\n", perr.Err, os.ErrExist)
		}
	}

	err = os.MkdirAll(filepath.Join(_testWD, "testdata/include/dir"), 0700)
	if err != nil {
		var perr *fs.PathError
		if !errors.As(err, &perr) {
			log.Fatal(err)
		}
		if !errors.Is(perr.Err, os.ErrExist) {
			log.Fatal(err)
		}
	}

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	afile := filepath.Join(_testWD, "testdata/index.html")

	cases := []struct {
		desc       string
		expErr     string
		expMapKeys []string
		opts       Options
	}{{
		desc:   "With empty dir",
		expErr: "open : no such file or directory",
	}, {
		desc: "With file",
		opts: Options{
			Root: afile,
		},
		expErr: fmt.Sprintf("New: Init: mount: createRoot: %s must be a directory", afile),
	}, {
		desc: "With directory",
		opts: Options{
			Root: filepath.Join(_testWD, "testdata"),
			Excludes: []string{
				"memfs_generate.go$",
				"direct$",
				"node_save$",
			},
		},
		expMapKeys: []string{
			"/",
			"/exclude",
			"/exclude/dir",
			"/exclude/index-link.css",
			"/exclude/index-link.html",
			"/exclude/index-link.js",
			"/include",
			"/include/dir",
			"/include/index.css",
			"/include/index.html",
			"/include/index.js",
			"/index.css",
			"/index.html",
			"/index.js",
			"/plain",
		},
	}, {
		desc: "With excludes",
		opts: Options{
			Root: filepath.Join(_testWD, "testdata"),
			Excludes: []string{
				`.*\.js$`,
				"memfs_generate.go$",
				"direct$",
				"node_save$",
			},
		},
		expMapKeys: []string{
			"/",
			"/exclude",
			"/exclude/dir",
			"/exclude/index-link.css",
			"/exclude/index-link.html",
			"/include",
			"/include/dir",
			"/include/index.css",
			"/include/index.html",
			"/index.css",
			"/index.html",
			"/plain",
		},
	}, {
		desc: "With includes",
		opts: Options{
			Root: filepath.Join(_testWD, "testdata"),
			Includes: []string{
				`.*\.js$`,
			},
			Excludes: []string{
				"memfs_generate.go$",
				"direct$",
				"node_save$",
			},
		},
		expMapKeys: []string{
			"/",
			"/exclude",
			"/exclude/dir",
			"/exclude/index-link.js",
			"/include",
			"/include/dir",
			"/include/index.js",
			"/index.js",
		},
	}}

	var (
		mfs *MemFS
		err error
	)
	for _, c := range cases {
		t.Log(c.desc)

		mfs, err = New(&c.opts)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		gotListNames := mfs.ListNames()
		test.Assert(t, "ListNames", c.expMapKeys, gotListNames)
	}
}

func TestMemFS_AddFile(t *testing.T) {
	cases := []struct {
		desc     string
		intPath  string
		extPath  string
		exp      *Node
		expError string
	}{{
		desc: "With empty internal path",
	}, {
		desc:     "With external path is not exist",
		intPath:  "internal/file",
		extPath:  "is/not/exist",
		expError: "AddFile: stat is/not/exist: no such file or directory",
	}, {
		desc:    "With file exist",
		intPath: "internal/file",
		extPath: "testdata/direct/add/file",
		exp: &Node{
			SysPath:     "testdata/direct/add/file",
			Path:        "internal/file",
			name:        "file",
			ContentType: "text/plain; charset=utf-8",
			size:        22,
			Content:     []byte("Test direct add file.\n"),
			GenFuncName: "generate_internal_file",
		},
	}, {
		desc:    "With directories exist",
		intPath: "internal/file2",
		extPath: "testdata/direct/add/file2",
		exp: &Node{
			SysPath:     "testdata/direct/add/file2",
			Path:        "internal/file2",
			name:        "file2",
			ContentType: "text/plain; charset=utf-8",
			size:        24,
			Content:     []byte("Test direct add file 2.\n"),
			GenFuncName: "generate_internal_file2",
		},
	}}

	opts := &Options{
		Root: "testdata",
	}
	mfs, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := mfs.AddFile(c.intPath, c.extPath)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error())
			continue
		}

		if got != nil {
			got.modTime = time.Time{}
			got.mode = 0
			got.Parent = nil
			got.Childs = nil
		}

		test.Assert(t, c.desc+": AddFile", c.exp, got)

		if c.exp == nil {
			continue
		}

		got, err = mfs.Get(c.intPath)
		if err != nil {
			t.Fatal(err)
		}

		if got != nil {
			got.modTime = time.Time{}
			got.mode = 0
			got.Parent = nil
			got.Childs = nil
		}

		test.Assert(t, c.desc+": Get", c.exp, got)
	}
}

func TestMemFS_Get(t *testing.T) {
	cases := []struct {
		expErr         string
		path           string
		expV           []byte
		expContentType []string
	}{{
		path: "/",
	}, {
		path: "/exclude",
	}, {
		path:   "/exclude/dir",
		expErr: os.ErrNotExist.Error(),
	}, {
		path:           "/exclude/index-link.css",
		expV:           []byte("body {\n}\n"),
		expContentType: []string{"text/css; charset=utf-8"},
	}, {
		path:           "/exclude/index-link.html",
		expV:           []byte("<html></html>\n"),
		expContentType: []string{"text/html; charset=utf-8"},
	}, {
		path: "/exclude/index-link.js",
		expV: []byte("function X() {}\n"),
		expContentType: []string{
			"text/javascript; charset=utf-8",
			"application/javascript",
		},
	}, {
		path: "/include",
	}, {
		path:   "/include/dir",
		expErr: os.ErrNotExist.Error(),
	}, {
		path:           "/include/index.css",
		expV:           []byte("body {\n}\n"),
		expContentType: []string{"text/css; charset=utf-8"},
	}, {
		path:           "/include/index.html",
		expV:           []byte("<html></html>\n"),
		expContentType: []string{"text/html; charset=utf-8"},
	}, {
		path: "/include/index.js",
		expV: []byte("function X() {}\n"),
		expContentType: []string{
			"text/javascript; charset=utf-8",
			"application/javascript",
		},
	}, {
		path:           "/index.css",
		expV:           []byte("body {\n}\n"),
		expContentType: []string{"text/css; charset=utf-8"},
	}, {
		path:           "/index.html",
		expV:           []byte("<html></html>\n"),
		expContentType: []string{"text/html; charset=utf-8"},
	}, {
		path: "/index.js",
		expContentType: []string{
			"text/javascript; charset=utf-8",
			"application/javascript",
		},
	}, {
		path:           "/plain",
		expContentType: []string{"application/octet-stream"},
	}, {
		path:   ``,
		expErr: `Get: empty path`,
	}}

	dir := filepath.Join(_testWD, "/testdata")

	opts := &Options{
		Root: dir,
		// Limit file size to allow testing Get from disk on file "index.js".
		MaxFileSize: 15,
	}

	mfs, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		got, err := mfs.Get(c.path)
		if err != nil {
			test.Assert(t, c.path+": error", c.expErr, err.Error())
			continue
		}

		if got.size <= opts.MaxFileSize {
			test.Assert(t, c.path+": node.Content", c.expV, got.Content)
		}

		if len(got.ContentType) == 0 && len(c.expContentType) == 0 {
			continue
		}

		found := false
		for _, expCT := range c.expContentType {
			if expCT == got.ContentType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expecting one of the Content-Type %v, got %s",
				c.expContentType, got.ContentType)
		}
	}
}

func TestMemFS_Get_refresh(t *testing.T) {
	type testCase struct {
		filePath string
	}

	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`internal/testdata/get_refresh_test.data`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		tempDir = t.TempDir()
		opts    = Options{
			Root:      tempDir,
			TryDirect: true,
		}

		mfs *MemFS
	)

	mfs, err = New(&opts)
	if err != nil {
		t.Fatal(err)
	}

	var listCase = []testCase{{
		filePath: `/dir-a/dir-b/file`,
	}, {
		filePath: `/dir-a/dir-b/file2`,
	}}

	var (
		c       testCase
		gotJSON bytes.Buffer
	)
	for _, c = range listCase {
		var fullpath = filepath.Join(tempDir, c.filePath)

		err = os.MkdirAll(filepath.Dir(fullpath), 0700)
		if err != nil {
			t.Fatal(err)
		}

		var expContent = tdata.Input[c.filePath]
		if len(expContent) != 0 {
			// Only create the file if content is set.
			err = os.WriteFile(fullpath, expContent, 0600)
			if err != nil {
				t.Fatal(err)
			}
		}

		// Try Get the file.

		var (
			tag      = c.filePath + `:error`
			expError = string(tdata.Output[tag])
			node     *Node
		)

		node, err = mfs.Get(c.filePath)
		if err != nil {
			test.Assert(t, tag, expError, err.Error())
			continue
		}

		// Check the tree of MemFS.

		var rawJSON []byte

		rawJSON, err = mfs.Root.JSON(9999, true, false)
		if err != nil {
			t.Fatal(err)
		}

		gotJSON.Reset()
		err = json.Indent(&gotJSON, rawJSON, ``, `  `)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, c.filePath+` content`, string(expContent), string(node.Content))

		var expJSON = string(tdata.Output[c.filePath])
		test.Assert(t, c.filePath+` JSON of memfs.Root`, expJSON, gotJSON.String())
	}
}

func TestMemFS_MarshalJSON(t *testing.T) {
	logp := "MarshalJSON"
	modTime := time.Date(2021, 7, 30, 20, 04, 00, 0, time.UTC)

	opts := &Options{
		Root: "testdata/direct/",
	}
	mfs, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}

	mfs.resetAllModTime(modTime)

	got, err := json.MarshalIndent(mfs, "", "\t")
	if err != nil {
		t.Fatal(err)
	}

	exp := `{
	"path": "/",
	"name": "/",
	"content_type": "",
	"mod_time": 1627675440,
	"mode_string": "drwxr-xr-x",
	"size": 0,
	"is_dir": true,
	"childs": [
		{
			"path": "/add",
			"name": "add",
			"content_type": "",
			"mod_time": 1627675440,
			"mode_string": "drwxr-xr-x",
			"size": 0,
			"is_dir": true,
			"childs": null
		}
	]
}`

	diffs := diff.Text([]byte(exp), got, diff.LevelLines)
	if len(diffs.Adds) != 0 {
		t.Fatalf("%s: adds: %v", logp, diffs.Adds)
	}
	if len(diffs.Dels) != 0 {
		t.Fatalf("%s: dels: %#v", logp, diffs.Dels)
	}
	if len(diffs.Changes) != 0 {
		t.Fatalf("%s: changes: %s", logp, diffs.Changes)
	}
}

func TestMemFS_RemoveChild(t *testing.T) {
	var (
		opts = &Options{
			Root:        `testdata`,
			MaxFileSize: -1,
		}
		mfs *MemFS
		err error
	)

	mfs, err = New(opts)
	if err != nil {
		t.Fatal(err)
	}

	var child = mfs.Root.Child(`plain`)
	if child == nil {
		t.Fatal(`Expecting child "plain", got nil`)
	}

	var nodeRemoved = mfs.RemoveChild(mfs.Root, child)
	if nodeRemoved == nil {
		t.Fatal(`Expecting child "plain", got nil`)
	}

	test.Assert(t, `RemoveChild`, child, nodeRemoved)

	child = mfs.Root.Child(`plain`)
	if child != nil {
		t.Fatalf(`Expecting child "plain" has been removed, got %v`, child)
	}
}

func TestMemFS_isIncluded(t *testing.T) {
	cases := []struct {
		desc    string
		inc     []string
		exc     []string
		sysPath []string
		exp     []bool
	}{{
		desc: "With empty includes and excludes",
		sysPath: []string{
			filepath.Join(_testWD, "/testdata"),
			filepath.Join(_testWD, "/testdata/index.html"),
		},
		exp: []bool{
			true,
			true,
		},
	}, {
		desc: "With excludes only",
		exc: []string{
			`.*/exclude`,
			`.*\.html$`,
		},
		sysPath: []string{
			filepath.Join(_testWD, "/testdata"),
			filepath.Join(_testWD, "/testdata/exclude"),
			filepath.Join(_testWD, "/testdata/exclude/dir"),
			filepath.Join(_testWD, "/testdata/include"),
			filepath.Join(_testWD, "/testdata"),
			filepath.Join(_testWD, "/testdata/index.html"),
			filepath.Join(_testWD, "/testdata/index.css"),
		},
		exp: []bool{
			true,
			false,
			false,
			true,
			true,
			false,
			true,
		},
	}, {
		desc: "With includes only",
		inc: []string{
			".*/include",
			`.*\.html$`,
		},
		sysPath: []string{
			filepath.Join(_testWD, "/testdata"),
			filepath.Join(_testWD, "/testdata/include"),
			filepath.Join(_testWD, "/testdata/include/dir"),
			filepath.Join(_testWD, "/testdata"),
			filepath.Join(_testWD, "/testdata/index.html"),
			filepath.Join(_testWD, "/testdata/index.css"),
		},
		exp: []bool{
			true,
			true,
			true,
			true,
			true,
			false,
		},
	}, {
		desc: "With excludes and includes",
		exc: []string{
			`.*/exclude`,
			`.*\.js$`,
		},
		inc: []string{
			`.*/include`,
			`.*\.(css|html)$`,
		},
		sysPath: []string{
			filepath.Join(_testWD, "/testdata"),
			filepath.Join(_testWD, "/testdata/index.html"),
			filepath.Join(_testWD, "/testdata/index.css"),

			filepath.Join(_testWD, "/testdata/exclude"),
			filepath.Join(_testWD, "/testdata/exclude/dir"),
			filepath.Join(_testWD, "/testdata/exclude/index-link.css"),
			filepath.Join(_testWD, "/testdata/exclude/index-link.html"),
			filepath.Join(_testWD, "/testdata/exclude/index-link.js"),

			filepath.Join(_testWD, "/testdata/include"),
			filepath.Join(_testWD, "/testdata/include/dir"),
			filepath.Join(_testWD, "/testdata/include/index.css"),
			filepath.Join(_testWD, "/testdata/include/index.html"),
			filepath.Join(_testWD, "/testdata/include/index.js"),
		},
		exp: []bool{
			true,
			true,
			true,

			false,
			false,
			false,
			false,
			false,

			true,
			true,
			true,
			true,
			false,
		},
	}}

	var (
		opts    *Options
		mfs     *MemFS
		fi      os.FileInfo
		sysPath string
		err     error
		x       int
		got     bool
	)
	for _, c := range cases {
		t.Log(c.desc)

		opts = &Options{
			Includes: c.inc,
			Excludes: c.exc,
		}
		mfs, err = New(opts)
		if err != nil {
			t.Fatal(err)
		}

		for x, sysPath = range c.sysPath {
			fi, err = os.Stat(sysPath)
			if err != nil {
				t.Fatal(err)
			}

			got = mfs.isExcluded(sysPath)
			if got {
				test.Assert(t, sysPath, !c.exp[x], got)
			} else {
				got = mfs.isIncluded(sysPath, fi)
				test.Assert(t, sysPath, c.exp[x], got)
			}
		}
	}
}

func TestScanDir(t *testing.T) {
	opts := Options{
		Root: "testdata/",
	}
	_, err := New(&opts)
	if err != nil {
		t.Fatal(err)
	}
}
