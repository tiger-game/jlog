// Copyright 2021 The tiger Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jlog

type Level int8

const (
	DEBUG Level = iota // just for developing stage
	INFO
	WARN
	ERROR

	MinLevel = DEBUG
	MaxLevel = ERROR + 1
)

var LevelNames = [MaxLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

var LevelFlags = [MaxLevel]byte{
	DEBUG: 'D',
	INFO:  'I',
	WARN:  'W',
	ERROR: 'E',
}

var LevelExtNames = [MaxLevel]string{
	DEBUG: "dbg",
	INFO:  "inf",
	WARN:  "wrn",
	ERROR: "err",
}

var NameLevels = map[string]Level{
	"DEBUG": DEBUG,
	"INFO":  INFO,
	"WARN":  WARN,
	"ERROR": ERROR,
}

func ExtentLevel(extName string) Level {
	for i, n := range LevelExtNames {
		if n == extName {
			return Level(i)
		}
	}
	return Level(-1)
}

// Get Level's string name
func (l Level) String() string {
	if l < MinLevel || l >= MaxLevel {
		return "Unknown Level"
	}
	return LevelNames[l]
}
