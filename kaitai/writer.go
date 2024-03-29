package kaitai

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// A Writer encapsulates writing binary data to files and memory.
type Writer struct {
	io.Writer
	buf [8]byte
}

// NewWriter creates and initializes a new Writer using w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{Writer: w}
}

// WriteU1 writes a uint8 to the underlying writer.
func (k *Writer) WriteU1(v uint8) error {
	k.buf[0] = v
	_, err := k.Write(k.buf[:1])
	if err != nil {
		return fmt.Errorf("WriteU1: failed to write uint8: %w", err)
	}
	return nil
}

// WriteU2be writes a uint16 in big-endian order to the underlying writer.
func (k *Writer) WriteU2be(v uint16) error {
	binary.BigEndian.PutUint16(k.buf[:2], v)
	_, err := k.Write(k.buf[:2])
	if err != nil {
		return fmt.Errorf("WriteU2be: failed to write uint16: %w", err)
	}
	return nil
}

// WriteU4be writes a uint32 in big-endian order to the underlying writer.
func (k *Writer) WriteU4be(v uint32) error {
	binary.BigEndian.PutUint32(k.buf[:4], v)
	_, err := k.Write(k.buf[:4])
	if err != nil {
		return fmt.Errorf("WriteU4be: failed to write uint32: %w", err)
	}
	return nil
}

// WriteU8be writes a uint64 in big-endian order to the underlying writer.
func (k *Writer) WriteU8be(v uint64) error {
	binary.BigEndian.PutUint64(k.buf[:8], v)
	_, err := k.Write(k.buf[:8])
	if err != nil {
		return fmt.Errorf("WriteU8be: failed to write uint64: %w", err)
	}
	return nil
}

// WriteU2le writes a uint16 in little-endian order to the underlying writer.
func (k *Writer) WriteU2le(v uint16) error {
	binary.LittleEndian.PutUint16(k.buf[:2], v)
	_, err := k.Write(k.buf[:2])
	if err != nil {
		return fmt.Errorf("WriteU2le: failed to write uint16: %w", err)
	}
	return nil
}

// WriteU4le writes a uint32 in little-endian order to the underlying writer.
func (k *Writer) WriteU4le(v uint32) error {
	binary.LittleEndian.PutUint32(k.buf[:4], v)
	_, err := k.Write(k.buf[:4])
	if err != nil {
		return fmt.Errorf("WriteU4le: failed to write uint32: %w", err)
	}
	return nil
}

// WriteU8le writes a uint64 in little-endian order to the underlying writer.
func (k *Writer) WriteU8le(v uint64) error {
	binary.LittleEndian.PutUint64(k.buf[:8], v)
	_, err := k.Write(k.buf[:8])
	if err != nil {
		return fmt.Errorf("WriteU8le: failed to write uint64: %w", err)
	}
	return nil
}

// WriteS1 writes an int8 to the underlying writer.
func (k *Writer) WriteS1(v int8) error {
	return k.WriteU1(uint8(v))
}

// WriteS2be writes an int16 in big-endian order to the underlying writer.
func (k *Writer) WriteS2be(v int16) error {
	return k.WriteU2be(uint16(v))
}

// WriteS4be writes an in32 in big-endian order to the underlying writer.
func (k *Writer) WriteS4be(v int32) error {
	return k.WriteU4be(uint32(v))
}

// WriteS8be writes an int64 in big-endian order to the underlying writer.
func (k *Writer) WriteS8be(v int64) error {
	return k.WriteU8be(uint64(v))
}

// WriteS2le writes an int16 in little-endian order to the underlying writer.
func (k *Writer) WriteS2le(v int16) error {
	return k.WriteU2le(uint16(v))
}

// WriteS4le writes an int32 in little-endian order to the underlying writer.
func (k *Writer) WriteS4le(v int32) error {
	return k.WriteU4le(uint32(v))
}

// WriteS8le writes an int64 in little-endian order to the underlying writer.
func (k *Writer) WriteS8le(v int64) error {
	return k.WriteU8le(uint64(v))
}

// WriteF4be writes a float32 in big-endian order to the underlying writer.
func (k *Writer) WriteF4be(v float32) error {
	return k.WriteU4be(math.Float32bits(v))
}

// WriteF8be writes a float64 in big-endian order to the underlying writer.
func (k *Writer) WriteF8be(v float64) error {
	return k.WriteU8be(math.Float64bits(v))
}

// WriteF4le writes a float32 in little-endian order to the underlying writer.
func (k *Writer) WriteF4le(v float32) error {
	return k.WriteU4le(math.Float32bits(v))
}

// WriteF8le writes a float64 in little-endian order to the underlying writer.
func (k *Writer) WriteF8le(v float64) error {
	return k.WriteU8le(math.Float64bits(v))
}

// WriteBytes writes the byte slice b to the underlying writer.
func (k *Writer) WriteBytes(b []byte) error {
	_, err := k.Write(b)
	if err != nil {
		return fmt.Errorf("WriteBytes: failed to write bytes: %w", err)
	}
	return nil
}
