// Copyright 2021 The tiger Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jlog

type Option func(l *logger)

func LogLevel(lv Level) Option {
	opt := func(l *logger) {
		l.file.level = lv
	}
	return opt
}

func LogDir(path string) Option {
	opt := func(l *logger) {
		l.file._InitLogPath(path)
	}
	return opt
}

// _LogDebug debug

func _LogDebug(debug bool) Option {
	opt := func(l *logger) {
		l.file.debug = debug
	}
	return opt
}


func _LogStd(std bool) Option {
	opt := func(l *logger) {
		l.std = std
	}
	return opt
}
