// Copyright 2021 The tiger Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !dev

package jlog

func StdLogInit() {
	stdLog = NewLogger(LogLevel(ERROR), _LogStd(true))
}

// GLogInit
// init gLog params.
func GLogInit(opts ...Option) {
	StdLogInit()
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
*/
func (l *logger) formatMsg(_ int) (string, int, bool) {
	return "", 0, false
}
