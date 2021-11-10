// Copyright 2021 The tiger Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jlog

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"
)

type logger struct {
	file       logFile
	logCh      chan logData
	pool       Pool
	std        bool
	closeWrite chan error
	waitClose  chan struct{}
	closed     chan struct{}
}

type logData struct {
	lv     Level
	file   string
	line   int
	debug  bool
	format func(buf *Buffer)
}

func NewLogger(opts ...Option) *logger {
	l := &logger{
		pool:       NewPool(),
		logCh:      make(chan logData, 1<<6),
		closeWrite: make(chan error, 1),
		waitClose:  make(chan struct{}),
		closed:     make(chan struct{}),
	}

	l.file.SetDefaultLevel()
	for _, opt := range opts {
		opt(l)
	}

	if l.IsNotCreateFile() {
		l.file._InitStdLog()
	}

	go func() {
		if err := l._GoLogger(); err != nil {
			l._CloseWithErr(err)
		}
		close(l.closed)
		if !l.std {
			l.file.Close()
		}
		close(l.waitClose)
	}()
	return l
}

func (l *logger) IsNotCreateFile() bool { return l.std && l.file.path == "" }

func (l *logger) Close() {
	close(l.closeWrite)
	<-l.waitClose
}

func (l *logger) _CloseWithErr(err error) {
	if err == nil {
		return
	}
	_StdLog().Errorf("logger Error:%v", err)
}

func (l *logger) _GoLogger() (err error) {
	var (
		data   logData
		ticker *time.Ticker
	)

	ticker = time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		if l.closeWrite == nil && len(l.logCh) == 0 {
			l.logCh = nil
			l.file.Flush()
			return
		}

		select {
		case data = <-l.logCh:
			buf := l.formatHeaderWithBodyFunction(data.lv, data.file, data.line, data.format, data.debug)
			if err = l.file.Write(data.lv, buf.Bytes(), l.IsNotCreateFile()); err != nil {
				buf.Free()
				return
			}
			buf.Free()
		case <-ticker.C:
			l.file.Flush()
		case err = <-l.closeWrite:
			l.closeWrite = nil
		}
	}
}

func DebugBufferAppend(buf *Buffer, arg interface{}) { appendArg2Buffer(buf, arg) }

func appendArg2Buffer(buf *Buffer, arg interface{}) {
	switch val := arg.(type) {
	case bool:
		buf.AppendBool(val)
	case int:
		buf.AppendInt(int64(val))
	case int8:
		buf.AppendInt(int64(val))
	case int16:
		buf.AppendInt(int64(val))
	case int32:
		buf.AppendInt(int64(val))
	case int64:
		buf.AppendInt(val)
	case uint:
		buf.AppendUint(uint64(val))
	case byte:
		_ = buf.WriteByte(val)
	case uint16:
		buf.AppendUint(uint64(val))
	case uint32:
		buf.AppendUint(uint64(val))
	case uint64:
		buf.AppendUint(val)
	case float32:
		buf.AppendFloat(float64(val), 32)
	case float64:
		buf.AppendFloat(val, 64)
	case []byte:
		_, _ = buf.Write(val)
	case string:
		_, _ = buf.WriteString(val)
	case error:
		_, _ = buf.WriteString(val.Error())
	default:
		appendValue2Buffer(buf, arg)
	}
}

func appendValue2Buffer(buf *Buffer, value interface{}) {
	switch val := reflect.ValueOf(value); val.Kind() {
	case reflect.Bool:
		buf.AppendBool(val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf.AppendInt(val.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		buf.AppendUint(val.Uint())
	case reflect.Float32:
		buf.AppendFloat(val.Float(), 32)
	case reflect.Float64:
		buf.AppendFloat(val.Float(), 64)
	case reflect.String:
		_, _ = buf.WriteString(val.String())
	default:
		enc := json.NewEncoder(buf)
		if err := enc.Encode(value); err != nil {
			_StdLog().Errorf("appendArg2Buffer default json marshal Error: %v", err)
			return
		}
	}
}

func (l *logger) Debug(args ...interface{}) { l.Output(DEBUG, "", 0, args...) }
func (l *logger) Info(args ...interface{})  { l.Output(INFO, "", 0, args...) }
func (l *logger) Warn(args ...interface{})  { l.Output(WARN, "", 0, args...) }
func (l *logger) Error(args ...interface{}) { l.Output(ERROR, "", 0, args...) }
func (l *logger) Debugf(format string, args ...interface{}) {
	l.Outputf(DEBUG, "", 0, format, args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.Outputf(INFO, "", 0, format, args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.Outputf(WARN, "", 0, format, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.Outputf(ERROR, "", 0, format, args...)
}

func (l *logger) Output(lv Level, prefix string, depth int, args ...interface{}) {
	if l == nil || !l.file._Check(lv) || len(args) == 0 {
		return
	}
	file, line, debug := l.formatMsg(depth)
	formatFunc := func(buf *Buffer) {
		if prefix != "" {
			_ = buf.WriteByte('[')
			_, _ = buf.WriteString(prefix)
			_ = buf.WriteByte(']')
		}
		for _, arg := range args {
			appendArg2Buffer(buf, arg)
			_ = buf.WriteByte(' ')
		}
	}

	select {
	case l.logCh <- logData{
		lv:     lv,
		file:   file,
		line:   line,
		format: formatFunc,
		debug:  debug,
	}:
	case <-l.closed:
		if l.std {
			return
		}
		buf := l.formatHeaderWithBodyFunction(lv, file, line, formatFunc, debug)
		if l.std {
			_, _ = fmt.Fprintf(os.Stderr, "logger discard: %s", buf.String())
		} else {
			_StdLog().Errorf("logger discard: %s", buf.String())
		}
		buf.Free()
	}
}

func (l *logger) Outputf(lv Level, prefix string, depth int, format string, args ...interface{}) {
	if l == nil || !l.file._Check(lv) {
		return
	}

	file, line, debug := l.formatMsg(depth)
	formatFunc := func(buf *Buffer) {
		if prefix != "" {
			_ = buf.WriteByte('[')
			_, _ = buf.WriteString(prefix)
			_ = buf.WriteByte(']')
		}
		_, _ = fmt.Fprintf(buf, format, args...)
	}
	select {
	case l.logCh <- logData{
		lv:     lv,
		file:   file,
		line:   line,
		format: formatFunc,
		debug:  debug,
	}:
	case <-l.closed:
		buf := l.formatHeaderWithBodyFunction(lv, file, line, formatFunc, debug)
		if l.std {
			_, _ = fmt.Fprintf(os.Stderr, "logger discard: %s", buf.String())
		} else {
			_StdLog().Errorf("logger discard: %s", buf.String())
		}
		buf.Free()
	}
}

// formatHeader formats a log header using the provided file name and line number.
func (l *logger) formatHeaderWithBodyFunction(lv Level, file string, line int, bodyFn func(buf *Buffer), dev bool) *Buffer {
	var (
		now time.Time
		buf *Buffer
		tmp [64]byte
	)

	now = time.Now()
	if line < 0 {
		line = 0 // not a real line number, but acceptable to someDigits
	}
	if lv > MaxLevel {
		lv = INFO // for safety.
	}

	// Avoid Fprintf, for speed. The format is so simple that we can do it quickly by hand.
	// It's worth about 3X. Fprintf is hard.
	year, month, day := now.Date()
	hour, minute, second := now.Clock()
	// yyyymmdd hh:mm:ss.uuuuuu [I] file:line
	buf = l.pool.Get()

	// header
	nDigits(tmp[:], 4, 0, year, '0')
	twoDigits(tmp[:], 4, int(month))
	twoDigits(tmp[:], 6, day)
	tmp[8] = ' '
	twoDigits(tmp[:], 9, hour)
	tmp[11] = ':'
	twoDigits(tmp[:], 12, minute)
	tmp[14] = ':'
	twoDigits(tmp[:], 15, second)
	tmp[17] = '.'
	nDigits(tmp[:], 6, 18, now.Nanosecond()/1000, '0')
	tmp[24] = ' '
	tmp[25] = '['
	tmp[26] = LevelFlags[lv]
	tmp[27] = ']'
	tmp[28] = ':'
	_, _ = buf.Write(tmp[:29])

	// body
	bodyFn(buf)

	// tail
	if dev {
		_, _ = buf.WriteString(" [")
		_, _ = buf.WriteString(file)
		tmp[0] = ':'
		n := someDigits(tmp[:], 1, line)
		tmp[n+1] = ']'
		tmp[n+2] = '\n'
		_, _ = buf.Write(tmp[:n+3])
	} else {
		_ = buf.WriteByte('\n')
	}

	return buf
}
