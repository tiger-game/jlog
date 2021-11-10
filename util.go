// Copyright 2021 The tiger Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jlog

import (
	"os"
)

// refer to: https://github.com/golang/glog
// Some custom tiny helper functions to print the log header efficiently.

const (
	timeFormat = "2006010215"
	digits     = "0123456789"
)

// twoDigits formats a zero-prefixed two-digit integer at buf.tmp[i].
func twoDigits(tmp []byte, i, d int) {
	tmp[i+1] = digits[d%10]
	d /= 10
	tmp[i] = digits[d%10]
}

// nDigits formats an n-digit integer at buf.tmp[i],
// padding with pad on the left.
// It assumes d >= 0.
func nDigits(tmp []byte, n, i, d int, pad byte) {
	j := n - 1
	for ; j >= 0 && d > 0; j-- {
		tmp[i+j] = digits[d%10]
		d /= 10
	}
	for ; j >= 0; j-- {
		tmp[i+j] = pad
	}
}

// someDigits formats a zero-prefixed variable-width integer at buf.tmp[i].
func someDigits(tmp []byte, i, d int) int {
	// Print into the top, then copy down. We know there's space for at least
	// a 10-digit number.
	j := len(tmp)
	for {
		j--
		tmp[j] = digits[d%10]
		d /= 10
		if d == 0 {
			break
		}
	}
	return copy(tmp[i:], tmp[j:])
}

func WithoutExt(path string) string {
	for i := len(path) - 1; i >= 0 && !os.IsPathSeparator(path[i]); i-- {
		if path[i] == '.' {
			return path[:i]
		}
	}
	return path
}
