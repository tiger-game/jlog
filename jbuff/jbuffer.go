// Copyright 2021 The tiger Authors. All rights reserved.

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// modify from go source code: src/bytes/buffer.go and "go.uber.org/zap/buffer"

package jbuff

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode/utf8"
	"unsafe"

	"google.golang.org/protobuf/proto"
)

// WARN: This File Only Use Little Endian Order For Binary!!!!

// smallBufferSize is an initial allocation minimal capacity.
const smallBufferSize = 64

const (
	BSize = 1 // byte, int8, *byte, *int8, bool, *bool
	WSize = 2 // int16, *int16, uint16, *uint16
	DSize = 4 // int32, *int32, uint32, *uint32
	QSize = 8 // int64, *int64, uint64, *uint64
)

// A JBuffer is a variable-sized buffer of bytes with Read and Write methods.
// The zero value for JBuffer is an empty buffer ready to use.
type JBuffer struct {
	buf []byte // contents are the bytes buf[off : len(buf)]
	off int    // read at &buf[off], write at &buf[len(buf)]
}

// ErrTooLarge is passed to panic if memory cannot be allocated to store data in a buffer.
var ErrTooLarge = errors.New("buffer.Buffer: too large")
var errNegativeRead = errors.New("buffer.Buffer: reader returned negative count from Read")

const maxInt = int(^uint(0) >> 1)

// Bytes returns a slice of length b.Len() holding the unread portion of the buffer.
// The slice is valid for use only until the next buffer modification (that is,
// only until the next call to a method like Read, Write, Reset, or Truncate).
// The slice aliases the buffer content at least until the next buffer modification,
// so immediate changes to the slice will affect the result of future reads.
func (j *JBuffer) Bytes() []byte { return j.buf[j.off:] }

// String returns the contents of the unread portion of the buffer
// as a string. If the JBuffer is a nil pointer, it returns "<nil>".
//
// To build strings more efficiently, see the strings.Builder type.
func (j *JBuffer) String() string {
	if j == nil {
		// Special case, useful in debugging.
		return "<nil>"
	}
	t := j.buf[j.off:]
	return *(*string)(unsafe.Pointer(&t))
	// return string(j.buf[j.off:])
}

// empty reports whether the unread portion of the buffer is empty.
func (j *JBuffer) empty() bool { return len(j.buf) <= j.off }

// Len returns the number of bytes of the unread portion of the buffer;
// b.Len() == len(b.Bytes()).
func (j *JBuffer) Len() int { return len(j.buf) - j.off }

// Cap returns the capacity of the buffer's underlying byte slice, that is, the
// total space allocated for the buffer's data.
func (j *JBuffer) Cap() int { return cap(j.buf) }

// Truncate discards all but the first n unread bytes from the buffer
// but continues to use the same allocated storage.
// It panics if n is negative or greater than the length of the buffer.
func (j *JBuffer) Truncate(n int) {
	if n == 0 {
		j.Reset()
		return
	}
	if n < 0 || n > j.Len() {
		panic("jbuff.JBuffer: truncation out of range")
	}
	j.buf = j.buf[:j.off+n]
}

// TrimNewline trims any final "\n" byte from the end of the buffer.
func (j *JBuffer) TrimNewline() {
	if i := len(j.buf) - 1; i >= 0 {
		if j.buf[i] == '\n' {
			j.buf = j.buf[:i]
		}
	}
}

// Reset resets the buffer to be empty,
// but it retains the underlying storage for use by future writes.
// Reset is the same as Truncate(0).
func (j *JBuffer) Reset() {
	j.buf = j.buf[:0]
	j.off = 0
}

// tryGrowByReslice is a inlineable version of grow for the fast-case where the
// internal buffer only needs to be resliced.
// It returns the index where bytes should be written and whether it succeeded.
func (j *JBuffer) tryGrowByReslice(n int) (int, bool) {
	if l := len(j.buf); n <= cap(j.buf)-l {
		j.buf = j.buf[:l+n]
		return l, true
	}
	return 0, false
}

// grow grows the buffer to guarantee space for n more bytes.
// It returns the index where bytes should be written.
// If the buffer can't grow it will panic with ErrTooLarge.
func (j *JBuffer) grow(n int) int {
	m := j.Len()
	// If buffer is empty, reset to recover space.
	if m == 0 && j.off != 0 {
		j.Reset()
	}
	// Try to grow by means of a reslice.
	if i, ok := j.tryGrowByReslice(n); ok {
		return i
	}
	if j.buf == nil && n <= smallBufferSize {
		j.buf = make([]byte, n, smallBufferSize)
		return 0
	}
	c := cap(j.buf)
	if n <= c/2-m {
		// We can slide things down instead of allocating a new
		// slice. We only need m+n <= c to slide, but
		// we instead let capacity get twice as large so we
		// don't spend all our time copying.
		copy(j.buf, j.buf[j.off:])
	} else if c > maxInt-c-n {
		panic(ErrTooLarge)
	} else {
		// Not enough space anywhere, we need to allocate.
		buf := makeSlice(2*c + n)
		copy(buf, j.buf[j.off:])
		j.buf = buf
	}
	// Restore j.off and len(j.buf).
	j.off = 0
	j.buf = j.buf[:m+n]
	return m
}

// Grow grows the buffer's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to the
// buffer without another allocation.
// If n is negative, Grow will panic.
// If the buffer can't grow it will panic with ErrTooLarge.
func (j *JBuffer) Grow(n int) {
	if n < 0 {
		panic("buffer.Buffer.Grow: negative count")
	}
	m := j.grow(n)
	j.buf = j.buf[:m]
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with ErrTooLarge.
func (j *JBuffer) Write(p []byte) (n int, err error) {
	m := j.Alloc(len(p))
	return copy(m, p), nil
}

// Writes not support type e.g. int,*int,[]int,[]*int,uint,*uint,[]uint,[]*uint and composite type
// the reason is that these types in different platform may have different size.
func (j *JBuffer) Writes(a ...interface{}) error {
	var (
		size int
		err  error
		s    int
	)

	// 1. calculate memory that store data.
	for i, arg := range a {
		if s, err = TypeSize(arg); err != nil {
			return fmt.Errorf("jbuff.JBuffer Writes type size the %dth data(%v) error: %w", i, arg, err)
		}
		size += s
	}

	// 2. maybe grow the memory.
	m := j.tryGrow(size)
	j.buf = j.buf[:m]

	// 3. write data to buffer.
	for i, arg := range a {
		if err = j.write(arg); err != nil {
			return fmt.Errorf("jbuff.JBuffer Writes the %dth data(%v) error: %w", i, arg, err)
		}
	}
	return nil
}

// FillTypeSize fill the type size to the buffer and return the start written index of the slice.
func (j *JBuffer) FillTypeSize(a interface{}) (int, error) {
	var (
		size int
		err  error
	)
	if size, err = TypeSize(a); err != nil {
		return 0, fmt.Errorf("buffers.Buffer WriteB error: %w", err)
	}
	m := j.tryGrow(size)
	return m, err
}

// tryGrow advance the slice length and return the start of write position.
// if enough memory, direct return or grow the memory and return.
func (j *JBuffer) tryGrow(n int) int {
	m, ok := j.tryGrowByReslice(n)
	if !ok {
		m = j.grow(n)
	}
	return m
}

// Alloc return n bytes memory in buffer.
func (j *JBuffer) Alloc(n int) []byte {
	m := j.tryGrow(n)
	return j.buf[m:]
}

// WriteTo writes data to w until the buffer is drained or an error occurs.
// The return value n is the number of bytes written; it always fits into an
// int, but it is int64 to match the io.WriterTo interface. Any error
// encountered during the write is also returned.
func (j *JBuffer) WriteTo(w io.Writer) (n int64, err error) {
	if nBytes := j.Len(); nBytes > 0 {
		m, e := w.Write(j.buf[j.off:])
		if m > nBytes {
			panic("buffer.Buffer.WriteTo: invalid Write count")
		}
		j.off += m
		n = int64(m)
		if e != nil {
			return n, e
		}
		// all bytes should have been written, by definition of
		// Write method in io.Writer
		if m != nBytes {
			return n, io.ErrShortWrite
		}
	}
	// Buffer is now empty; reset.
	j.Reset()
	return n, nil
}

// WriteRune write the UTF-8 encoding of Unicode code point r to the
// buffer, returning its length and an error, which is always nil but is
// included to match bufio.Writer's WriteRune. The buffer is grown as needed;
// if it becomes too large, WriteRune will panic with ErrTooLarge.
func (j *JBuffer) WriteRune(r rune) (n int, err error) {
	if r < utf8.RuneSelf {
		_ = j.WriteByte(byte(r))
		return 1, nil
	}
	m := j.tryGrow(utf8.UTFMax)
	n = utf8.EncodeRune(j.buf[m:m+utf8.UTFMax], r)
	j.buf = j.buf[:m+n]
	return n, nil
}

// WriteString write the contents of s to the buffer, growing the buffer as
// needed. The return value n is the length of s; err is always nil. If the
// buffer becomes too large, WriteString will panic with ErrTooLarge.
func (j *JBuffer) WriteString(s string) (n int, err error) {
	m := j.Alloc(len(s))
	return copy(m, s), nil
}

// WriteInt64 write an integer(64bits) binary to the buffer.
func (j *JBuffer) WriteInt64(i int64) { j.WriteUint64(uint64(i)) }

// WriteUint64 write an unsigned integer(64bits) binary to the buffer.
func (j *JBuffer) WriteUint64(i uint64) {
	m := j.Alloc(QSize)
	Uint64ToBytes(i, m)
}

// WriteVar write an unsigned integer(64bits) binary to the buffer.
// WARN: only write to unsigned integer
func (j *JBuffer) WriteVar(i uint64) {
	m := j.Alloc(Size(i))
	VarToBytes(i, m)
}

// WriteInt32 write an integer(32bits) binary to the buffer.
func (j *JBuffer) WriteInt32(i int32) { j.WriteUint32(uint32(i)) }

// WriteUint32 write an unsigned integer(32bits) binary to the buffer.
func (j *JBuffer) WriteUint32(i uint32) {
	m := j.Alloc(DSize)
	Uint32ToBytes(i, m)
}

// WriteInt16 write an integer(16bits) binary to the buffer.
func (j *JBuffer) WriteInt16(i int16) { j.WriteUint16(uint16(i)) }

// WriteUint16 write an unsigned integer(16bits) binary to the buffer.
func (j *JBuffer) WriteUint16(i uint16) {
	m := j.Alloc(WSize)
	Uint16ToBytes(i, m)
}

// WriteByte (alias WriteUint8) write the byte c to the buffer, growing the buffer as needed.
// The returned error is always nil, but is included to match bufio.Writer's
// WriteByte. If the buffer becomes too large, WriteByte will panic with
// ErrTooLarge.
func (j *JBuffer) WriteByte(c byte) error {
	m := j.Alloc(BSize)
	m[0] = c
	return nil
}

// WriteInt8 write an integer(8bits) binary to the buffer.
func (j *JBuffer) WriteInt8(i int8) { _ = j.WriteByte(byte(i)) }

// WriteBool write a bool binary to the buffer.
func (j *JBuffer) WriteBool(v bool) {
	var i byte
	if v {
		i = 1
	}
	_ = j.WriteByte(i)
}

// WriteFloat64 write a float64(64bits) binary to the buffer.
func (j *JBuffer) WriteFloat64(f float64) { j.WriteUint64(*(*uint64)(unsafe.Pointer(&f))) }

// WriteFloat32 write a float32(32bits) binary to the buffer.
func (j *JBuffer) WriteFloat32(f float32) { j.WriteUint32(*(*uint32)(unsafe.Pointer(&f))) }

func (j *JBuffer) Marshal(p proto.Message) error {
	var (
		size int
		err  error
	)
	size = proto.Size(p)
	if size == 0 {
		return nil
	}
	m := j.Alloc(size)
	_, err = proto.MarshalOptions{}.MarshalAppend(m, p)
	return err
}

// AppendInt appends an integer to the underlying buffer (assuming base 10).
func (j *JBuffer) AppendInt(i int64) { j.buf = strconv.AppendInt(j.buf, i, 10) }

// AppendUint appends an unsigned integer to the underlying buffer (assuming
// base 10).
func (j *JBuffer) AppendUint(i uint64) { j.buf = strconv.AppendUint(j.buf, i, 10) }

// AppendBool appends a bool to the underlying buffer.
func (j *JBuffer) AppendBool(v bool) { j.buf = strconv.AppendBool(j.buf, v) }

// AppendFloat appends a float to the underlying buffer. It doesn't quote NaN
// or +/- Inf.
func (j *JBuffer) AppendFloat(f float64, bitSize int) {
	j.buf = strconv.AppendFloat(j.buf, f, 'f', -1, bitSize)
}

// MinRead is the minimum slice size passed to a Read call by
// Buffer.ReadFrom. As long as the JBuffer has at least MinRead bytes beyond
// what is required to hold the contents of r, ReadFrom will not grow the
// underlying buffer.
const MinRead = 512

// ReadFrom reads data from r until EOF and appends it to the buffer, growing
// the buffer as needed. The return value n is the number of bytes read. Any
// error except io.EOF encountered during the read is also returned. If the
// buffer becomes too large, ReadFrom will panic with ErrTooLarge.
func (j *JBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	for {
		i := j.grow(MinRead)
		j.buf = j.buf[:i]
		m, e := r.Read(j.buf[i:cap(j.buf)])
		if m < 0 {
			panic(errNegativeRead)
		}

		j.buf = j.buf[:i+m]
		n += int64(m)
		if e == io.EOF {
			return n, nil // e is EOF, so return nil explicitly
		}
		if e != nil {
			return n, e
		}
	}
}

// makeSlice allocates a slice of size n. If the allocation fails, it panics
// with ErrTooLarge.
func makeSlice(n int) []byte {
	// If the make fails, give a known error.
	defer func() {
		if recover() != nil {
			panic(ErrTooLarge)
		}
	}()
	return make([]byte, n)
}

// Read reads the next len(p) bytes from the buffer or until the buffer
// is drained. The return value n is the number of bytes read. If the
// buffer has no data to return, err is io.EOF (unless len(p) is zero);
// otherwise it is nil.
func (j *JBuffer) Read(p []byte) (n int, err error) {
	if j.empty() {
		// Buffer is empty, reset to recover space.
		j.Reset()
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n = copy(p, j.buf[j.off:])
	j.off += n
	return n, nil
}

func (j *JBuffer) Reads(a ...interface{}) (err error) {
	for _, arg := range a {
		if err = j.read(arg); err != nil {
			return err
		}
	}
	return nil
}

func (j *JBuffer) Unmarshal(p proto.Message) error {
	var (
		err error
	)
	if err = proto.Unmarshal(j.Bytes(), p); err != nil {
		return err
	}
	j.off += proto.Size(p)
	return nil
}

// Next returns a slice containing the next n bytes from the buffer,
// advancing the buffer as if the bytes had been returned by Read.
// If there are fewer than n bytes in the buffer, Next returns the entire buffer.
// The slice is only valid until the next call to a read or write method.
func (j *JBuffer) Next(n int) []byte {
	m := j.Len()
	if n > m {
		n = m
	}
	data := j.buf[j.off : j.off+n]
	j.off += n
	return data
}

// ReadByte (alias ReadUint8) reads and returns the next byte from the buffer.
// If no byte is available, it returns error io.EOF.
func (j *JBuffer) ReadByte() (byte, error) {
	if j.empty() {
		// Buffer is empty, reset to recover space.
		j.Reset()
		return 0, io.EOF
	}
	c := j.buf[j.off]
	j.off++
	return c, nil
}

// ReadInt8 reads and returns an integer(8bits) from the buffer(binary).
// If no byte is available, it returns error io.EOF.
func (j *JBuffer) ReadInt8() (int8, error) {
	v, err := j.ReadByte()
	return int8(v), err
}

// ReadInt64 reads and returns an integer(64bits) from the buffer(binary).
// If no byte is available, it returns error io.EOF.
func (j *JBuffer) ReadInt64() (int64, error) {
	v, err := j.ReadUint64()
	return int64(v), err
}

// ReadUint64 reads and returns an integer(64bits) from the buffer(binary).
// If no byte is available, it returns error io.EOF.
func (j *JBuffer) ReadUint64() (uint64, error) {
	if j.empty() {
		// Buffer is empty, reset to recover space.
		j.Reset()
		return 0, io.EOF
	}
	v := Bytes2Uint64(j.buf[j.off:])
	j.off += QSize
	return v, nil
}

// ReadVarint reads and returns an integer(64bits) from the buffer(binary).
// If no byte is available, it returns error io.EOF.
func (j *JBuffer) ReadVarint() (uint64, error) {
	if j.empty() {
		// Buffer is empty, reset to recover space.
		j.Reset()
		return 0, io.EOF
	}
	v, n := Bytes2Var(j.buf[j.off:])
	if n == -1 {
		return 0, fmt.Errorf("jbuff.JBuffer read varint error: data truncated")
	}
	j.off += n
	return v, nil
}

// ReadInt32 reads and returns an integer(32bits) from the buffer(binary).
// If no byte is available, it returns error io.EOF.
func (j *JBuffer) ReadInt32() (int32, error) {
	v, err := j.ReadUint32()
	return int32(v), err
}

// ReadUint32 reads and returns an integer(32bits) from the buffer(binary).
// If no byte is available, it returns error io.EOF.
func (j *JBuffer) ReadUint32() (uint32, error) {
	if j.empty() {
		// Buffer is empty, reset to recover space.
		j.Reset()
		return 0, io.EOF
	}
	v := Bytes2Uint32(j.buf[j.off:])
	j.off += DSize
	return v, nil
}

// ReadInt16 reads and returns an integer(16bits) from the buffer(binary).
// If no byte is available, it returns error io.EOF.
func (j *JBuffer) ReadInt16() (int16, error) {
	v, err := j.ReadUint16()
	return int16(v), err
}

// ReadUint16 reads and returns an integer(16bits) from the buffer((binary)).
// If no byte is available, it returns error io.EOF.
func (j *JBuffer) ReadUint16() (uint16, error) {
	if j.empty() {
		// Buffer is empty, reset to recover space.
		j.Reset()
		return 0, io.EOF
	}
	v := Bytes2Uint16(j.buf[j.off:])
	j.off += WSize
	return v, nil
}

// ReadFloat64 reads and returns an integer(64bits) from the buffer(binary).
// If no byte is available, it returns error io.EOF.
func (j *JBuffer) ReadFloat64() (float64, error) {
	if j.empty() {
		// Buffer is empty, reset to recover space.
		j.Reset()
		return 0, io.EOF
	}
	v := Bytes2Float64(j.buf[j.off:])
	j.off += QSize
	return v, nil
}

// ReadFloat32 reads and returns an integer(32bits) from the buffer(binary).
// If no byte is available, it returns error io.EOF.
func (j *JBuffer) ReadFloat32() (float32, error) {
	if j.empty() {
		// Buffer is empty, reset to recover space.
		j.Reset()
		return 0, io.EOF
	}
	v := Bytes2Float32(j.buf[j.off:])
	j.off += DSize
	return v, nil
}

// ReadBool reads and returns an bool from the buffer(binary).
// If no byte is available, it returns error io.EOF.
func (j *JBuffer) ReadBool() (bool, error) {
	v, err := j.ReadByte()
	return v > 0, err
}

// ReadRune reads and returns the next UTF-8-encoded
// Unicode code point from the buffer.
// If no bytes are available, the error returned is io.EOF.
// If the bytes are an erroneous UTF-8 encoding, it
// consumes one byte and returns U+FFFD, 1.
func (j *JBuffer) ReadRune() (r rune, size int, err error) {
	if j.empty() {
		// Buffer is empty, reset to recover space.
		j.Reset()
		return 0, 0, io.EOF
	}
	c := j.buf[j.off]
	if c < utf8.RuneSelf {
		j.off++
		return rune(c), 1, nil
	}
	r, n := utf8.DecodeRune(j.buf[j.off:])
	j.off += n
	return r, n, nil
}

// ReadBytes reads until the first occurrence of delim in the input,
// returning a slice containing the data up to and including the delimiter.
// If ReadBytes encounters an error before finding a delimiter,
// it returns the data read before the error and the error itself (often io.EOF).
// ReadBytes returns err != nil if and only if the returned data does not end in
// delim.
func (j *JBuffer) ReadBytes(delim byte) (line []byte, err error) {
	slice, err := j.readSlice(delim)
	// return a copy of slice. The buffer's backing array may
	// be overwritten by later calls.
	line = append(line, slice...)
	return line, err
}

// readSlice is like ReadBytes but returns a reference to internal buffer data.
func (j *JBuffer) readSlice(delim byte) (line []byte, err error) {
	i := bytes.IndexByte(j.buf[j.off:], delim)
	end := j.off + i + 1
	if i < 0 {
		end = len(j.buf)
		err = io.EOF
	}
	line = j.buf[j.off:end]
	j.off = end
	return line, err
}

// ReadString reads until the first occurrence of delim in the input,
// returning a string containing the data up to and including the delimiter.
// If ReadString encounters an error before finding a delimiter,
// it returns the data read before the error and the error itself (often io.EOF).
// ReadString returns err != nil if and only if the returned data does not end
// in delim.
func (j *JBuffer) ReadString(delim byte) (line string, err error) {
	slice, err := j.readSlice(delim)
	return string(slice), err
}

func (j *JBuffer) CopyR(d, s int) error {
	if s >= d || s > j.Len() || d < j.off {
		return fmt.Errorf("jbuff.JBuffer Copy parameter error: dst(%v), src(%v), len(%v)", d, s, j.Len())
	}

	step := d - s
	j.Alloc(step)
	for i := j.Len() - 1; i >= d; i-- {
		j.buf[i] = j.buf[i-step]
	}

	return nil
}

func (j *JBuffer) Copy(d, s int) error {
	if d >= s || d > j.Len() || s < j.off {
		return fmt.Errorf("jbuff.JBuffer Copy parameter error: dst(%v), src(%v), len(%v)", d, s, j.Len())
	}
	copy(j.buf[d:], j.buf[s:])
	return nil
}

// NewJBuffer creates and initializes a new JBuffer using buf as its
// initial contents. The new JBuffer takes ownership of buf, and the
// caller should not use buf after this call. NewJBuffer is intended to
// prepare a JBuffer to read existing data. It can also be used to set
// the initial size of the internal buffer for writing. To do that,
// buf should have the desired capacity but a length of zero.
//
// In most cases, new(JBuffer) (or just declaring a JBuffer variable) is
// sufficient to initialize a JBuffer.
func NewJBuffer(buf []byte) *JBuffer { return &JBuffer{buf: buf} }

// NewJBufferString creates and initializes a new JBuffer using string s as its
// initial contents. It is intended to prepare a buffer to read an existing
// string.
//
// In most cases, new(JBuffer) (or just declaring a JBuffer variable) is
// sufficient to initialize a JBuffer.
func NewJBufferString(s string) *JBuffer {
	return &JBuffer{buf: []byte(s)}
}
