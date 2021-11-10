// Copyright 2021 The tiger Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jlog

import (
	"sync"

	"github.com/tiger-game/jlog/jbuff"
)

const _size = 1024

type Buffer struct {
	jbuff.JBuffer
	pool Pool
}

func (b *Buffer) Free() { b.pool.put(b) }

func NewBuffer() *Buffer {
	b := &Buffer{}
	b.Grow(_size)
	return b
}

// A Pool is a type-safe wrapper around a sync.Pool.
type Pool struct {
	p *sync.Pool
}

// NewPool constructs a new Pool.
func NewPool() Pool {
	return Pool{p: &sync.Pool{
		New: func() interface{} {
			return NewBuffer()
		},
	}}
}

// Get retrieves a Buffer from the pool, creating one if necessary.
func (p Pool) Get() *Buffer {
	buf := p.p.Get().(*Buffer)
	buf.Reset()
	buf.pool = p
	return buf
}

func (p Pool) put(buf *Buffer) { p.p.Put(buf) }
