// Copyright 2021 The tiger Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jlog

var gLog *logger
var stdLog *logger

// GLog
// fetch gLog
func GLog() Logger        { return gLog }
func _StdLog() Logger     { return stdLog }
func DebugStdLog() Logger { return stdLog }

func CloseGLog() {
	gLog.Close()
	CloseStdLog()
}

func CloseStdLog() {
	stdLog.Close()
}

// Debug
// global gLog for debug
func Debug(args ...interface{}) { gLog.Output(DEBUG, "GLog", 0, args...) }

// Info
// global gLog for info
func Info(args ...interface{}) { gLog.Output(INFO, "GLog", 0, args...) }

// Warn
// global gLog for warn
func Warn(args ...interface{}) { gLog.Output(WARN, "GLog", 0, args...) }

// Error
// global gLog for error
func Error(args ...interface{}) { gLog.Output(ERROR, "GLog", 0, args...) }

// Debugf
// global gLog for debug
func Debugf(format string, args ...interface{}) {
	gLog.Outputf(DEBUG, "GLog", 0, format, args...)
}

// Infof
// global gLog for info
func Infof(format string, args ...interface{}) {
	gLog.Outputf(INFO, "GLog", 0, format, args...)
}

// Warnf
// global gLog for warn
func Warnf(format string, args ...interface{}) {
	gLog.Outputf(WARN, "GLog", 0, format, args...)
}

// Errorf
// global gLog for error
func Errorf(format string, args ...interface{}) {
	gLog.Outputf(ERROR, "GLog", 0, format, args...)
}

type CustomLogger struct {
	*logger
	prefix string
	level  Level
}

func (cl *CustomLogger) _ControlFlag(lv Level) bool {
	return cl.logger != nil && cl.level >= lv
}

func (cl *CustomLogger) Debug(args ...interface{}) {
	if !cl._ControlFlag(DEBUG) {
		return
	}
	cl.Output(DEBUG, cl.prefix, 0, args...)
}
func (cl *CustomLogger) Info(args ...interface{}) {
	if !cl._ControlFlag(INFO) {
		return
	}
	cl.Output(INFO, cl.prefix, 0, args...)
}
func (cl *CustomLogger) Warn(args ...interface{}) {
	if !cl._ControlFlag(WARN) {
		return
	}
	cl.Output(WARN, cl.prefix, 0, args...)
}
func (cl *CustomLogger) Error(args ...interface{}) {
	if !cl._ControlFlag(ERROR) {
		return
	}
	cl.Output(ERROR, cl.prefix, 0, args...)
}
func (cl *CustomLogger) Debugf(format string, args ...interface{}) {
	if !cl._ControlFlag(DEBUG) {
		return
	}
	cl.Outputf(DEBUG, cl.prefix, 0, format, args...)
}
func (cl *CustomLogger) Infof(format string, args ...interface{}) {
	if !cl._ControlFlag(INFO) {
		return
	}
	cl.Outputf(INFO, cl.prefix, 0, format, args...)
}
func (cl *CustomLogger) Warnf(format string, args ...interface{}) {
	if !cl._ControlFlag(WARN) {
		return
	}
	cl.Outputf(WARN, cl.prefix, 0, format, args...)
}
func (cl *CustomLogger) Errorf(format string, args ...interface{}) {
	if !cl._ControlFlag(ERROR) {
		return
	}
	cl.Outputf(ERROR, cl.prefix, 0, format, args...)
}

func NewLogByPrefixLevel(prefix string, level Level) Logger {
	ul := &CustomLogger{
		prefix: prefix,
		logger: gLog,
		level:  level,
	}
	return ul
}

func NewLogByPrefix(prefix string) Logger {
	ul := &CustomLogger{
		prefix: prefix,
		logger: gLog,
		level:  ERROR,
	}
	return ul
}
