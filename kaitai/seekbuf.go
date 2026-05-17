package kaitai

import (
	"errors"
	"io"
)

// SeekableBuffer is an in-memory buffer that implements io.ReadWriteSeeker.
type SeekableBuffer struct {
	data []byte
	pos  int64
}

// NewSeekableBuffer creates a SeekableBuffer with the given initial data.
func NewSeekableBuffer(data []byte) *SeekableBuffer {
	return &SeekableBuffer{data: data}
}

// NewSeekableBufferSize creates a SeekableBuffer pre-allocated to size bytes.
func NewSeekableBufferSize(size int) *SeekableBuffer {
	return &SeekableBuffer{data: make([]byte, size)}
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
	end := sb.pos + int64(len(p))
	if end > int64(len(sb.data)) {
		// Grow buffer
		if end > int64(cap(sb.data)) {
			newData := make([]byte, end, end*2)
			copy(newData, sb.data)
			sb.data = newData
		} else {
			sb.data = sb.data[:end]
		}
	}
	n = copy(sb.data[sb.pos:], p)
	sb.pos += int64(n)
	return n, nil
}

func (sb *SeekableBuffer) Seek(offset int64, whence int) (int64, error) {
	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = sb.pos + offset
	case io.SeekEnd:
		newPos = int64(len(sb.data)) + offset
	default:
		return 0, errors.New("SeekableBuffer.Seek: invalid whence")
	}
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
