// Copyright 2021 The tiger Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build dev

package jlog

import (
	"runtime"
	"strings"
)

func StdLogInit() {
	stdLog = NewLogger(_LogDebug(true), LogLevel(ERROR), _LogStd(true))
}

// init gLog params.
func GLogInit(opts ...Option) {
	StdLogInit()
	opts = append(opts, _LogDebug(true))
	gLog = NewLogger(opts...)
}

/*
header formats a log header as defined by the C++ implementation.
It returns a buffer containing the formatted header and the user's file and line number.
The depth specifies how many stack frames above lives the source line to be identified in the log message.
Log lines have this form:
	yyyymmdd hh:mm:ss.uuuuuu [L] msg... [file:line]
where the fields are defined as follows:
	L                A single character, representing the log level (eg 'I' for INFO)
	yyyy             The year (zero padded)
	mm               The month (zero padded; ie May is '05')
	dd               The day (zero padded)
	hh:mm:ss.uuuuuu  Time in hours, minutes and fractional seconds
	msg              The user-supplied message
	file             The file name
	line             The line number
*/
func (l *logger) formatMsg(depth int) (string, int, bool) {
	_, file, line, ok := runtime.Caller(3 + depth)
	if !ok {
		file = "???"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	return file, line, true
	// return l.formatHeaderWithBodyFunction(lv, file, line, bodyFn, true)
}
