// Copyright 2021 The tiger Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jlog

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const MaxSize = 1e9 // 1GB

type FileStream struct {
	rawFile    *os.File
	writer     *bufio.Writer
	writeSize  int
	createTime int64
	idx        int
}

func (f *FileStream) IsWriter() bool        { return f.writer != nil && f.rawFile != nil }
func (f *FileStream) OverflowMaxSize() bool { return f.writeSize >= MaxSize }
func (f *FileStream) RotateByTime() bool {
	nowNano := time.Now().UnixNano()
	if f.createTime/int64(time.Hour) >= nowNano/int64(time.Hour) {
		return false
	}
	if f.createTime != 0 {
		f.idx = 0
	}
	f.createTime = nowNano
	return true
}

func (f *FileStream) _Init(raw *os.File) {
	f.rawFile = raw
	f.writer = bufio.NewWriter(raw)
}

func (f *FileStream) Write(p []byte) error {
	if !f.IsWriter() {
		return nil
	}
	n, err := f.writer.Write(p)
	f.writeSize += n
	return err
}

func (f *FileStream) Flush() error {
	if !f.IsWriter() {
		return fmt.Errorf("FileStream Flush Error: f.rawFile == nil(%v), f.writer == nil(%v)", f.rawFile == nil, f.writer == nil)
	}
	return f.writer.Flush()
}

func (f *FileStream) SymLink(dst string) {
	var err error
	if err = os.RemoveAll(dst); err != nil {
		_StdLog().Errorf("_Dup2 Remove Error: %v", err)
		return
	}

	if err = os.Symlink(filepath.Base(f.rawFile.Name()), dst); err != nil {
		_StdLog().Errorf("_Dup2 Symlink Error: %v", err)
	}
}

func (f *FileStream) Close() {
	if !f.IsWriter() {
		return
	}

	var err error
	if err = f.Flush(); err != nil {
		_StdLog().Errorf("FileStream Close Error: Flush %v", err)
	}
	if err = f.rawFile.Close(); err != nil {
		_StdLog().Errorf("FileStream Close Error: Close %v", err)
	}
	f.writer = nil
	f.rawFile = nil
	f.writeSize = 0
	f.idx++
}
