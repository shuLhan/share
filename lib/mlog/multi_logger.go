// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mlog

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

//
// MultiLogger represent a single log writer that write to multiple outputs.
// MultiLogger can have zero or more writers for standard output (normal log)
// and zero or more writers for standard error.
//
// Each call to write APIs (Errf, Fatalf, or Outf) will be prefixed with
// time format in UTC and optional prefix.
//
type MultiLogger struct {
	bufPool    *sync.Pool
	timeFormat string
	prefix     []byte
	outs       map[string]NamedWriter
	errs       map[string]NamedWriter
	qout       chan []byte
	qerr       chan []byte
	qflush     chan bool
	qerrFlush  chan bool
	qoutFlush  chan bool
}

//
// NewMultiLogger create and initialize new MultiLogger.
//
func NewMultiLogger(timeFormat, prefix string, outs, errs []NamedWriter) (mlog *MultiLogger) {
	mlog = &MultiLogger{
		bufPool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
		timeFormat: timeFormat,
		prefix:     []byte(prefix),
		outs:       make(map[string]NamedWriter),
		errs:       make(map[string]NamedWriter),
		qout:       make(chan []byte, 512),
		qerr:       make(chan []byte, 512),
		qerrFlush:  make(chan bool, 1),
		qoutFlush:  make(chan bool, 1),
		qflush:     make(chan bool, 1),
	}
	for _, w := range outs {
		name := w.Name()
		if len(name) == 0 {
			continue
		}
		mlog.outs[name] = w
	}
	for _, w := range errs {
		name := w.Name()
		if len(name) == 0 {
			continue
		}
		mlog.errs[name] = w
	}

	go mlog.processErrorQueue()
	go mlog.processOutputQueue()
	return mlog
}

//
// Errf write the formatted string and its optional values to all error
// writers.
//
// The format string does not end with new line, it will be added.
//
func (mlog *MultiLogger) Errf(format string, v ...interface{}) {
	if len(mlog.errs) == 0 {
		return
	}
	mlog.writeTo(mlog.qerr, format, v...)
}

//
// Fatalf is equal to Errf and os.Exit(1).
//
func (mlog *MultiLogger) Fatalf(format string, v ...interface{}) {
	mlog.Errf(format, v...)
	mlog.Flush()
	os.Exit(1)
}

//
// Flush all writes and wait until it finished.
//
func (mlog *MultiLogger) Flush() {
	mlog.qerrFlush <- true
	mlog.qoutFlush <- true
	<-mlog.qflush
	<-mlog.qflush
}

//
// Outf write the formatted string and its values to all output writers.
//
func (mlog *MultiLogger) Outf(format string, v ...interface{}) {
	if len(mlog.outs) == 0 {
		return
	}
	mlog.writeTo(mlog.qout, format, v...)
}

//
// RegisterErrorWriter register the named writer to one of error writers.
// The writer Name() must not be empty or it will not registered.
//
func (mlog *MultiLogger) RegisterErrorWriter(errw NamedWriter) {
	name := errw.Name()
	if len(name) == 0 {
		return
	}
	mlog.errs[name] = errw
}

//
// RegisterOutputWriter register the named writer to one of output writers.
// The writer Name() must not be empty or it will not registered.
//
func (mlog *MultiLogger) RegisterOutputWriter(outw NamedWriter) {
	name := outw.Name()
	if len(name) == 0 {
		return
	}
	mlog.outs[name] = outw
}

//
// SetPrefix set the default prefix for the subsequence writes.
//
func (mlog *MultiLogger) SetPrefix(prefix string) {
	mlog.prefix = []byte(prefix)
}

//
// SetTimeFormat set the default time format for the subsequence writes.
//
func (mlog *MultiLogger) SetTimeFormat(layout string) {
	mlog.timeFormat = layout
}

//
// UnregisterErrorWriter remove the error writer by name.
//
func (mlog *MultiLogger) UnregisterErrorWriter(name string) {
	delete(mlog.errs, name)
}

//
// UnregisterOutputWriter remove the output writer by name.
//
func (mlog *MultiLogger) UnregisterOutputWriter(name string) {
	delete(mlog.outs, name)
}

func (mlog *MultiLogger) processErrorQueue() {
	var err error
	for {
		select {
		case b := <-mlog.qerr:
			if len(b) == 0 {
				b = append(b, '\n')
			} else if b[len(b)-1] != '\n' {
				b = append(b, '\n')
			}
			for _, w := range mlog.errs {
				_, err = w.Write(b)
				if err != nil {
					log.Printf("MultiLogger: %s: %s", w.Name(), err)
				}
			}
		case <-mlog.qerrFlush:
			for x := 0; x < len(mlog.qerr); x++ {
				b := <-mlog.qerr
				if len(b) == 0 {
					b = append(b, '\n')
				} else if b[len(b)-1] != '\n' {
					b = append(b, '\n')
				}
				for _, w := range mlog.errs {
					_, err = w.Write(b)
					if err != nil {
						log.Printf("MultiLogger: %s: %s", w.Name(), err)
					}
				}
			}
			mlog.qflush <- true
		}
	}
}

func (mlog *MultiLogger) processOutputQueue() {
	var err error
	for {
		select {
		case b := <-mlog.qout:
			for _, w := range mlog.outs {
				_, err = w.Write(b)
				if err != nil {
					log.Printf("MultiLogger: %s: %s", w.Name(), err)
				}
			}
		case <-mlog.qoutFlush:
			for x := 0; x < len(mlog.qout); x++ {
				b := <-mlog.qout
				for _, w := range mlog.outs {
					_, err = w.Write(b)
					if err != nil {
						log.Printf("MultiLogger: %s: %s", w.Name(), err)
					}

				}
			}
			mlog.qflush <- true
		}
	}
}

func (mlog *MultiLogger) writeTo(q chan []byte, format string, v ...interface{}) {
	var (
		buf    = mlog.bufPool.Get().(*bytes.Buffer)
		bufFmt = mlog.bufPool.Get().(*bytes.Buffer)
		args   = make([]interface{}, 0, len(v)+2)
	)
	buf.Reset()
	bufFmt.Reset()

	if len(mlog.timeFormat) > 0 {
		args = append(args, time.Now().UTC().Format(mlog.timeFormat))
		bufFmt.WriteString("%s ")
	}
	if len(mlog.prefix) > 0 {
		args = append(args, mlog.prefix)
		bufFmt.WriteString("%s ")
	}
	bufFmt.WriteString(format)
	args = append(args, v...)
	fmt.Fprintf(buf, bufFmt.String(), args...)

	q <- libbytes.Copy(buf.Bytes())

	mlog.bufPool.Put(bufFmt)
	mlog.bufPool.Put(buf)
}
