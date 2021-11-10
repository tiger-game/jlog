// Copyright 2021 The tiger Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jlog

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tiger-game/jlog/utils"
)

const DefaultLoggerLevel = INFO

type logFile struct {
	streams [MaxLevel]FileStream
	level   Level
	path    string
	logName string
	debug   bool
}

func (l *logFile) _InitLogPath(path string) {
	l.path = path
	if err := os.MkdirAll(l.path, 0774); err != nil {
		_StdLog().Errorf("LoggerFile CreateDir Error: %v", err)
	}
	l.logName = WithoutExt(filepath.Base(os.Args[0]))
	l._CheckFileIndex()
}

func (l *logFile) _InitStdLog() {
	var rawFile *os.File
	for lv := range LevelExtNames {
		if !l._Check(Level(lv)) {
			continue
		}
		rawFile = os.Stdout
		if lv == int(ERROR) {
			rawFile = os.Stderr
		}
		l.streams[lv]._Init(rawFile)
	}
}

func (l *logFile) SetDefaultLevel() {
	// set default level for gLog
	if l.level < DefaultLoggerLevel {
		l.level = DefaultLoggerLevel
	}
}

func (l *logFile) _CheckFileIndex() {
	timeStr := time.Now().Format(timeFormat)

	_ = filepath.WalkDir(l.path, func(path string, d fs.DirEntry, err error) error {
		if d == nil || d.IsDir() {
			return nil
		}
		idx, lv := parseIdxAndLv(timeStr, d.Name())
		if lv == Level(-1) {
			return nil
		}
		if l.streams[lv].idx >= idx {
			return nil
		}
		l.streams[lv].idx = idx
		if info, err := d.Info(); err == nil {
			l.streams[lv].writeSize = int(info.Size())
		}
		return nil
	})
}

func parseIdxAndLv(timeStr, fname string) (int, Level) {
	var i int
	if !strings.Contains(fname, timeStr) {
		return -1, Level(-1)
	}

	if i = strings.LastIndexByte(fname, '.'); i == -1 {
		return -1, Level(-1)
	}
	fname = fname[:i]
	if i = strings.LastIndexByte(fname, '.'); i == -1 {
		return -1, Level(-1)
	}
	l := ExtentLevel(fname[i+1:])
	fname = fname[:i]
	if i = strings.LastIndexByte(fname, '.'); i == -1 {
		return -1, Level(-1)
	}
	idx, err := utils.Str2Int(fname[i+1:])
	if err != nil {
		return -1, Level(-1)
	}
	return idx, l
}

func (l *logFile) _Check(lv Level) bool {
	return (lv == DEBUG && l.debug) || (lv != DEBUG && lv <= l.level)
}

func (l *logFile) _NewCreateFile(lv Level) {
	var (
		err     error
		rawFile *os.File
	)

	l.streams[lv].Close()
	if rawFile, err = os.OpenFile(l._DirFileName(lv), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664); err != nil {
		_StdLog().Errorf("NewFile Error: %v", err)
		return
	}
	l.streams[lv]._Init(rawFile)
	l.streams[lv].SymLink(l._RedirectFile(lv))
}

func (l *logFile) _DirFileName(lv Level) string {
	var (
		timeStr = time.Now().Format(timeFormat)
		ext     = LevelExtNames[lv]
	)
	buffer := bytes.NewBuffer(make([]byte, 0, len(l.logName)+len(timeStr)+len(ext)+6))
	// format: /dir/../appName_200601021504_inf.log
	buffer.WriteString(l.logName)
	buffer.WriteByte('.')
	buffer.WriteString(timeStr)
	buffer.WriteByte('.')
	buffer.WriteString(strconv.Itoa(l.streams[lv].idx))
	buffer.WriteByte('.')
	buffer.WriteString(ext)
	buffer.WriteString(".log")
	return filepath.Join(l.path, buffer.String())
}

func (l *logFile) _RedirectFile(lv Level) string {
	ext := LevelExtNames[lv]
	// format: logName.inf
	buffer := bytes.NewBuffer(make([]byte, 0, len(l.logName)+len(ext)+1))
	buffer.WriteString(l.logName)
	buffer.WriteByte('.')
	buffer.WriteString(ext)
	return filepath.Join(l.path, buffer.String())
}

func (l *logFile) Close() {
	for _, f := range l.streams {
		f.Close()
	}
}

func (l *logFile) Write(level Level, data []byte, notCreateFile bool) (err error) {
	for lv := DEBUG; lv <= level; lv++ {
		if notCreateFile && lv != level {
			continue
		}
		if !l._Check(lv) {
			continue
		}

		if !notCreateFile && (l.streams[lv].OverflowMaxSize() || l.streams[lv].RotateByTime()) {
			l._NewCreateFile(lv)
		}
		if err = l.streams[lv].Write(data); err != nil {
			_StdLog().Errorf("logFile Write Error: %v", err)
			return err
		}
	}
	return
}

func (l *logFile) Flush() {
	for lv := DEBUG; lv < MaxLevel; lv++ {
		if !l.streams[lv].IsWriter() {
			continue
		}

		if err := l.streams[lv].Flush(); err != nil {
			_StdLog().Errorf("logFile Flush Error: %v", err)
		}
	}
}
