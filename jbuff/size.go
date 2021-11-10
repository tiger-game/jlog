// Copyright 2021 The tiger Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jbuff

import (
	"fmt"
)

func TypeSize(data interface{}) (int, error) {
	var (
		size int
		err  error
	)
	switch dataImp := data.(type) {
	case bool, int8, uint8, *bool, *int8, *uint8:
		size = BSize
	case []bool:
		size = len(dataImp)
	case []int8:
		size = len(dataImp)
	case []uint8:
		size = len(dataImp)
	case int16, uint16, *int16, *uint16:
		size = WSize
	case []int16:
		size = WSize * len(dataImp)
	case []uint16:
		size = WSize * len(dataImp)
	case int32, uint32, *int32, *uint32:
		size = DSize
	case []int32:
		size = DSize * len(dataImp)
	case []uint32:
		size = DSize * len(dataImp)
	case int64, uint64, *int64, *uint64:
		size = QSize
	case []int64:
		size = QSize * len(dataImp)
	case []uint64:
		size = QSize * len(dataImp)
	case string:
		size = len(dataImp)
	default:
		err = fmt.Errorf("buffers.Buffer TypeSize error: not support data type:%T", data)
	}
	return size, err
}

// WriteAt write the basic data into the start m position
func (j *JBuffer) WriteAt(a interface{}, m int) error {
	var (
		size int
		err  error
	)

	if m < j.off {
		return fmt.Errorf("buffers.Buffer error: index(%v) too large to write data, write start pos(%v)", m, j.off)
	}

	if size, err = TypeSize(a); err != nil {
		return fmt.Errorf("buffers.Buffer error: %w", err)
	}

	if m >= len(j.buf) || m+size > len(j.buf) {
		return fmt.Errorf("buffers.Buffer error: index(%v) data size(%v) too large to write data, write start pos(%v)", m, size, j.off)
	}

	switch d := a.(type) {
	case int8:
		j.buf[m] = uint8(d)
	case uint8:
		j.buf[m] = d
	case *int8:
		j.buf[m] = uint8(*d)
	case *uint8:
		j.buf[m] = *d
	case int16:
		binaryOrder.PutUint16(j.buf[m:], uint16(d))
	case uint16:
		binaryOrder.PutUint16(j.buf[m:], d)
	case *int16:
		binaryOrder.PutUint16(j.buf[m:], uint16(*d))
	case *uint16:
		binaryOrder.PutUint16(j.buf[m:], *d)
	case int32:
		binaryOrder.PutUint32(j.buf[m:], uint32(d))
	case uint32:
		binaryOrder.PutUint32(j.buf[m:], d)
	case *int32:
		binaryOrder.PutUint32(j.buf[m:], uint32(*d))
	case *uint32:
		binaryOrder.PutUint32(j.buf[m:], *d)
	case int64:
		binaryOrder.PutUint64(j.buf[m:], uint64(d))
	case uint64:
		binaryOrder.PutUint64(j.buf[m:], d)
	case *int64:
		binaryOrder.PutUint64(j.buf[m:], uint64(*d))
	case *uint64:
		binaryOrder.PutUint64(j.buf[m:], *d)
	default:
		return fmt.Errorf("jbuff.JBuffer error: not suport data type(%T)", d)
	}
	return nil
}

// write appends the basic data to the buffer.
func (j *JBuffer) write(a interface{}) error {
	var err error
	switch v := a.(type) {
	case *bool:
		j.WriteBool(*v)
	case bool:
		j.WriteBool(v)
	case []bool:
		for _, x := range v {
			j.WriteBool(x)
		}
	case *int8:
		j.WriteInt8(*v)
	case int8:
		j.WriteInt8(v)
	case []int8:
		for _, x := range v {
			j.WriteInt8(x)
		}
	case *uint8:
		_ = j.WriteByte(*v)
	case uint8:
		_ = j.WriteByte(v)
	case []uint8: // equal []byte
		_, err = j.Write(v)
	case *int16:
		j.WriteInt16(*v)
	case int16:
		j.WriteInt16(v)
	case []int16:
		for _, x := range v {
			j.WriteInt16(x)
		}
	case *uint16:
		j.WriteUint16(*v)
	case uint16:
		j.WriteUint16(v)
	case []uint16:
		for _, x := range v {
			j.WriteUint16(x)
		}
	case *int32:
		j.WriteInt32(*v)
	case int32:
		j.WriteInt32(v)
	case []int32:
		for _, x := range v {
			j.WriteInt32(x)
		}
	case *uint32:
		j.WriteUint32(*v)
	case uint32:
		j.WriteUint32(v)
	case []uint32:
		for _, x := range v {
			j.WriteUint32(x)
		}
	case *int64:
		j.WriteInt64(*v)
	case int64:
		j.WriteInt64(v)
	case []int64:
		for _, x := range v {
			j.WriteInt64(x)
		}
	case *uint64:
		j.WriteUint64(*v)
	case uint64:
		j.WriteUint64(v)
	case []uint64:
		for _, x := range v {
			j.WriteUint64(x)
		}
	case string:
		_, err = j.WriteString(v)
	default:
		err = fmt.Errorf("buffers.Buffer write error: not support data type, %T", v)
	}
	return err
}

// read the data to the basic parameter interface variable.
func (j *JBuffer) read(a interface{}) error {
	var err error
	switch d := a.(type) {
	case *bool:
		*d, err = j.ReadBool()
	case *int8:
		*d, err = j.ReadInt8()
	case *uint8:
		*d, err = j.ReadByte()
	case *int16:
		*d, err = j.ReadInt16()
	case *uint16:
		*d, err = j.ReadUint16()
	case *int32:
		*d, err = j.ReadInt32()
	case *uint32:
		*d, err = j.ReadUint32()
	case *int64:
		*d, err = j.ReadInt64()
	case *uint64:
		*d, err = j.ReadUint64()
	default:
		err = fmt.Errorf("buffers.Buffer read error: not support data type(%T)", a)
	}
	return err
}
