// Copyright 2021 The tiger Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utils

import (
	"strconv"
	"unsafe"
)

// Str2Bytes converts string to byte slice without a memory allocation.
func Str2Bytes(s string) (b []byte) {
	bytes := struct {
		string
		cap int
	}{s, len(s)}
	return *(*[]byte)(unsafe.Pointer(&bytes))
}

// Bytes2Str converts byte slice to string without a memory allocation.
func Bytes2Str(b []byte) string { return *(*string)(unsafe.Pointer(&b)) }

func Str2Int(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	i, err := strconv.ParseInt(s, 10, 64)
	return int(i), err
}

func Str2Int16(s string) (int16, error) {
	i, err := strconv.ParseInt(s, 10, 16)
	return int16(i), err
}

func Str2Uint16(s string) uint16 {
	i, _ := strconv.ParseUint(s, 10, 16)
	return uint16(i)
}

func Str2Float32(s string) (float32, error) {
	if s == "" {
		return 0, nil
	}
	f, err := strconv.ParseFloat(s, 32)
	return float32(f), err
}

func Str2Bool(s string) (bool, error) {
	if s == "" {
		return false, nil
	}
	return strconv.ParseBool(s)
}
