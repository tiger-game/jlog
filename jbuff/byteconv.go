// Copyright 2021 The tiger Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jbuff

import (
	"encoding/binary"
	"math/bits"
	"unsafe"
)

var (
	binaryOrder = binary.LittleEndian
)

func Byte2Bool(b byte) bool        { return b > 0 }
func Byte2Int8(b byte) int8        { return int8(b) }
func Bytes2Int16(b []byte) int16   { return int16(Bytes2Uint16(b)) }
func Bytes2Uint16(b []byte) uint16 { return binaryOrder.Uint16(b) }
func Bytes2Int32(b []byte) int32   { return int32(Bytes2Uint32(b)) }
func Bytes2Uint32(b []byte) uint32 { return binaryOrder.Uint32(b) }
func Bytes2Int64(b []byte) int64   { return int64(Bytes2Uint64(b)) }
func Bytes2Uint64(b []byte) uint64 { return binaryOrder.Uint64(b) }
func Bytes2Var(b []byte) (v uint64, n int) {
	for i := 0; i < len(b); i++ {
		val := uint64(b[i])
		v += val << (i * 7)
		if (i != 9 && val < 0x80) || (i == 9 && val < 2) {
			return v, i + 1
		}
		v -= 0x80 << (i * 7)
	}
	return 0, -1
}
func Bytes2Float32(b []byte) float32 {
	i := Bytes2Uint32(b)
	return *(*float32)(unsafe.Pointer(&i))
}
func Bytes2Float64(b []byte) float64 {
	i := Bytes2Uint64(b)
	return *(*float64)(unsafe.Pointer(&i))
}

func Bool2Byte(i bool) byte {
	v := byte(0)
	if i {
		v = 1
	}
	return v
}

func Int8ToByte(i int8) byte           { return byte(i) }
func Int16ToBytes(i int16, b []byte)   { Uint16ToBytes(uint16(i), b) }
func Uint16ToBytes(i uint16, b []byte) { binaryOrder.PutUint16(b, i) }
func Int32ToBytes(i int32, b []byte)   { Uint32ToBytes(uint32(i), b) }
func Uint32ToBytes(i uint32, b []byte) { binaryOrder.PutUint32(b, i) }
func Int64ToBytes(i int64, b []byte)   { Uint64ToBytes(uint64(i), b) }
func Uint64ToBytes(i uint64, b []byte) { binaryOrder.PutUint64(b, i) }
func VarToBytes(v uint64, b []byte) {
	var cnt int
	for cnt = 1; cnt <= 9; cnt++ {
		if v < 1<<(7*cnt) {
			break
		}
	}

	for i := 0; i < cnt; i++ {
		if i != cnt-1 {
			b[i] = byte((v>>(7*i))&0x7f | 0x80)
		} else if i != 9 {
			b[i] = byte(v >> (7 * i))
		} else {
			b[i] = 1
		}
	}
}
func Float64ToBytes(f float64, b []byte) {
	i := *(*uint64)(unsafe.Pointer(&f))
	Uint64ToBytes(i, b)
}

func Float32ToBytes(f float32, b []byte) {
	i := *(*uint32)(unsafe.Pointer(&f))
	Uint32ToBytes(i, b)
}

// Size returns the encoded size of a varint.
// The size is guaranteed to be within 1 and 10, inclusive.
func Size(v uint64) int {
	// This computes 1 + (bits.Len64(v)-1)/7.
	// 9/64 is a good enough approximation of 1/7
	return int(9*uint32(bits.Len64(v))+64) / 64
}
