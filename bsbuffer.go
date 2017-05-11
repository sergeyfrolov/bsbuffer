// Copyright 2017 Sergey Frolov
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package bsbuffer

import (
	"bytes"
	"io"
	"sync"
)

// B - Blocking - Read() calls are blocking.
// S - Safe - Supports arbitrary #s of readers and writers.
// Buffer,
type BSBuffer struct {
	sync.Mutex
	buf        bytes.Buffer
	r          io.PipeReader
	w          io.PipeWriter

	readerWait chan bool
	hasData chan bool
	writerWait chan bool
}

func NewBSBuffer(initialSize int) *BSBuffer {
	bsb := new(BSBuffer)

	b := make([]byte, initialSize)
	bsb.buf = bytes.NewBuffer(b)

	bsb.r, bsb.w = io.Pipe()

	bsb.hasData = make(chan bool, 1)
	go bsb.engine()
	return bsb
}

func (b *BSBuffer) engine() {
	for {
		select {
		case _ = <-b.hasData:
			b.Lock()
			if b.buf.Len() == 0 {
				// empty buffer, but b.hasData -- it was closed
				b.r.Close()
				b.w.Close()
				return
			}
			slice := b.buf.Bytes()
			sliceCopy := make([]byte, len(slice))
			copy(sliceCopy, slice)
			b.Unlock()
			b.w.Write(sliceCopy)
			continue
		}
	}
}


func (b *BSBuffer) Read(p []byte) (n int, err error) {
	n, err = b.r.Read(p)
	return
}

func (b *BSBuffer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	b.Lock()
	n, err = b.buf.Write(p)
	select {
	case <-b.hasData:
	default:
	}
	b.Unlock()
	return
}

// Second close is unexpected behaviour (probably panic)
func (b *BSBuffer) Close() {
	b.Lock()
	close(b.hasData)
	b.Unlock()
}
