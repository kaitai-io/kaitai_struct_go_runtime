package kaitai

import (
	"errors"
	"fmt"
	"io"
	"math"
)

// ErrBufferCapacity indicates that a write or seek would exceed its bounded or
// representable capacity. This includes fixed buffers, fixed-size fields written
// with WriteBytesLimit, and growable buffers whose requested range cannot be
// represented or allocated.
var ErrBufferCapacity = errors.New("fixed buffer capacity exceeded")

// SeekableBuffer is an in-memory buffer that implements io.ReadWriteSeeker.
// A rejected write preserves both its contents and current position.
type SeekableBuffer struct {
	data  []byte
	pos   int64
	fixed bool
}

// NewSeekableBuffer creates a SeekableBuffer with the given initial data.
func NewSeekableBuffer(data []byte) *SeekableBuffer {
	return &SeekableBuffer{data: data}
}

// NewSeekableBufferSize creates a SeekableBuffer pre-allocated to size bytes.
func NewSeekableBufferSize(size int) *SeekableBuffer {
	return &SeekableBuffer{data: make([]byte, size)}
}

// NewFixedSeekableBuffer creates a zero-filled buffer of exactly size bytes.
// Unlike NewSeekableBufferSize, it cannot grow: writes or seeks beyond its end
// return ErrBufferCapacity without modifying its contents or position.
// As with NewSeekableBufferSize, a negative size panics.
func NewFixedSeekableBuffer(size int) *SeekableBuffer {
	return &SeekableBuffer{data: make([]byte, size), fixed: true}
}

func (sb *SeekableBuffer) Read(p []byte) (n int, err error) {
	if sb.pos >= int64(len(sb.data)) {
		return 0, io.EOF
	}
	n = copy(p, sb.data[sb.pos:])
	sb.pos += int64(n)
	return n, nil
}

func (sb *SeekableBuffer) Write(p []byte) (n int, err error) {
	writeLen := int64(len(p))
	if sb.pos > math.MaxInt64-writeLen {
		return 0, fmt.Errorf("SeekableBuffer.Write: requested range overflows: %w", ErrBufferCapacity)
	}
	end64 := sb.pos + writeLen
	if sb.fixed && end64 > int64(len(sb.data)) {
		return 0, fmt.Errorf("SeekableBuffer.Write: requested range exceeds fixed capacity: %w", ErrBufferCapacity)
	}
	maxInt := int64(^uint(0) >> 1)
	if end64 > maxInt {
		return 0, fmt.Errorf("SeekableBuffer.Write: requested range is not representable: %w", ErrBufferCapacity)
	}
	end := int(end64)
	if end > len(sb.data) {
		// Grow buffer
		if end > cap(sb.data) {
			newData, ok := growSeekableBuffer(sb.data, end)
			if !ok {
				return 0, fmt.Errorf("SeekableBuffer.Write: requested range cannot be allocated: %w", ErrBufferCapacity)
			}
			sb.data = newData
		} else {
			sb.data = sb.data[:end]
		}
	}
	n = copy(sb.data[sb.pos:], p)
	sb.pos += int64(n)
	return n, nil
}

func growSeekableBuffer(data []byte, end int) (grown []byte, ok bool) {
	defer func() {
		if recover() != nil {
			grown = nil
			ok = false
		}
	}()
	newCap := end
	maxInt := int(^uint(0) >> 1)
	if end <= maxInt/2 {
		newCap = end * 2
	}
	grown = make([]byte, end, newCap)
	copy(grown, data)
	return grown, true
}

func (sb *SeekableBuffer) Seek(offset int64, whence int) (int64, error) {
	var base int64
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		base = sb.pos
	case io.SeekEnd:
		base = int64(len(sb.data))
	default:
		return 0, errors.New("SeekableBuffer.Seek: invalid whence")
	}
	if sb.fixed && offset > int64(len(sb.data))-base {
		return 0, ErrBufferCapacity
	}
	newPos := base + offset
	if newPos < 0 {
		return 0, errors.New("SeekableBuffer.Seek: negative position")
	}
	sb.pos = newPos
	return newPos, nil
}

// Bytes returns the buffer contents.
func (sb *SeekableBuffer) Bytes() []byte {
	return sb.data
}

// Len returns the current length of the buffer.
func (sb *SeekableBuffer) Len() int {
	return len(sb.data)
}

func (sb *SeekableBuffer) remainingWriteCapacity() (int64, bool) {
	if !sb.fixed {
		return 0, false
	}
	return int64(len(sb.data)) - sb.pos, true
}
