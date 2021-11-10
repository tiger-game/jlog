// Copyright 2021 The tiger Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jbuff_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tiger-game/tiger/jbuff"
)

var _ = Describe("JBuffer", func() {
	Describe("Buffer", func() {
		_ = jbuff.NewJBufferString("")
		var buffer = &jbuff.JBuffer{}

		BeforeEach(func() {
			buffer.Reset()
		})

		It("show number in human style", func() {
			buffer.AppendInt(12345)
			Expect(buffer.String()).To(Equal("12345"))
			buffer.AppendBool(true)
			Expect(buffer.String()).To(Equal("12345true"))
			buffer.AppendFloat(1.2, 32)
			Expect(buffer.String()).To(Equal("12345true1.2"))
			buffer.AppendFloat(1.23456e10, 64)
			Expect(buffer.String()).To(Equal("12345true1.212345600000"))
		})

		It("show number in binary style", func() {
			buffer.WriteInt32(-12345)
			Expect(buffer.Bytes()).To(Equal([]byte{0xc7, 0xcf, 0xff, 0xff}))
			buffer.WriteUint64(9876543210)
			Expect(buffer.Bytes()).To(Equal([]byte{0xc7, 0xcf, 0xff, 0xff, 0xea, 0x16, 0xb0, 0x4c, 0x02, 0x00, 0x00, 0x00}))

			i32, _ := buffer.ReadInt32()
			Expect(i32).To(Equal(int32(-12345)))

			u64, _ := buffer.ReadUint64()
			Expect(u64).To(Equal(uint64(9876543210)))
		})

		It("show mix number", func() {
			buffer.WriteInt32(-12345)
			Expect(buffer.Bytes()).To(Equal([]byte{0xc7, 0xcf, 0xff, 0xff}))
			buffer.AppendInt(12345)
			Expect(buffer.Bytes()).To(Equal([]byte{0xc7, 0xcf, 0xff, 0xff, '1', '2', '3', '4', '5'}))
			buffer.AppendBool(true)
			Expect(buffer.Bytes()).To(Equal([]byte{0xc7, 0xcf, 0xff, 0xff, '1', '2', '3', '4', '5', 't', 'r', 'u', 'e'}))
			buffer.AppendFloat(1.2, 32)
			Expect(buffer.Bytes()).To(Equal([]byte{0xc7, 0xcf, 0xff, 0xff, '1', '2', '3', '4', '5', 't', 'r', 'u', 'e', '1', '.', '2'}))
			buffer.WriteUint64(9876543210)
			Expect(buffer.Bytes()).To(Equal([]byte{0xc7, 0xcf, 0xff, 0xff, '1', '2', '3', '4', '5', 't', 'r', 'u', 'e', '1', '.', '2', 0xea, 0x16, 0xb0, 0x4c, 0x02, 0x00, 0x00, 0x00}))
		})

		It("writeAt", func() {
			buffer.WriteInt32(-12345)
			m, _ := buffer.FillTypeSize(int32(12))
			_, _ = buffer.WriteString("Hello!")
			_ = buffer.WriteAt(int32(256), m)
			Expect(buffer.Bytes()).To(Equal([]byte{0xc7, 0xcf, 0xff, 0xff, 0x00, 0x01, 0x00, 0x00, 'H', 'e', 'l', 'l', 'o', '!'}))
		})

		It("CopyR", func() {
			for i := 0; i < 10; i++ {
				buffer.WriteUint16(uint16(i))
			}

			_ = buffer.CopyR(10, 8)
			_ = buffer.WriteAt(uint16(9), 8)
			Expect(buffer.Bytes()).To(Equal([]byte{0x00, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x09, 0x00, 0x04, 0x00, 0x05, 0x00, 0x06, 0x00, 0x07, 0x00, 0x08, 0x00, 0x09, 0x00}))

			_ = buffer.CopyR(2, 0)
			_ = buffer.WriteAt(uint16(3), 0)
			Expect(buffer.Bytes()).To(Equal([]byte{0x03, 0x00, 0x00, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x09, 0x00, 0x04, 0x00, 0x05, 0x00, 0x06, 0x00, 0x07, 0x00, 0x08, 0x00, 0x09, 0x00}))
		})
	})

	Describe("Varints", func() {
		cases := []struct {
			x uint64
			n int
		}{
			{0, 1},
			{1<<7 - 1, 1},
			{1 << 7, 2},
			{1<<8 - 1, 2}, // max uint8
			{1 << 8, 2},
			{1<<14 - 1, 2},
			{1 << 14, 3},
			{1<<16 - 1, 3}, // max uint16
			{1 << 16, 3},
			{1<<21 - 1, 3},
			{1 << 21, 4},
			{1<<28 - 1, 4},
			{1 << 28, 5},
			{1<<32 - 1, 5}, // max uint32
			{1 << 32, 5},
			{1<<35 - 1, 5},
			{1 << 35, 6},
			{1<<42 - 1, 6},
			{1 << 42, 7},
			{1<<49 - 1, 7},
			{1 << 49, 8},
			{1<<56 - 1, 8},
			{1 << 56, 9},
			{1<<63 - 1, 9},
			{1 << 63, 10},
			{1<<64 - 1, 10}, // max uint64
		}
		It("Varints convert", func() {
			for _, c := range cases {
				Expect(jbuff.Size(c.x)).To(Equal(c.n))
				slc := make([]byte, c.n)
				jbuff.VarToBytes(c.x, slc)
				val, cnt := jbuff.Bytes2Var(slc)
				Expect(val).To(Equal(c.x))
				Expect(cnt).To(Equal(c.n))
			}
		})

	})
})
