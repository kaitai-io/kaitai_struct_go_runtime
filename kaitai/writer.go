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

	bits     uint64
	bitsLeft int
	bitsLe   bool
	// bitsReserved reports that the underlying cursor is already one byte
	// past bits. This happens when a bounded bit write fails its atomic
	// capacity preflight: the prior pending byte remains logically pending,
	// but is materialized so the physical position matches Pos().
	bitsReserved bool

	children []*ChildWriter
}

// NewWriter creates and initializes a new Writer using w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{Writer: w}
}

// Write aligns pending bits before writing p. The returned count excludes
// the alignment byte, if any.
func (k *Writer) Write(p []byte) (int, error) {
	err := k.AlignToByte()
	if err != nil {
		return 0, err
	}
	return k.writeNotAligned(p)
}

// Pos returns the position after any pending partial byte, if the stream is
// seekable. It does not flush pending bits.
func (k *Writer) Pos() (int64, error) {
	switch w := k.Writer.(type) {
	case io.Seeker:
		n, err := w.Seek(0, io.SeekCurrent)
		if err != nil {
			return 0, fmt.Errorf("Pos: failed to get pos: %w", err)
		}
		if k.bitsLeft > 0 && !k.bitsReserved {
			n++
		}
		return n, nil
	default:
		return 0, fmt.Errorf("Pos: %w", ErrWriterNotSeekable)
	}
}

// HasPendingBits reports whether a partial byte is waiting to be completed.
// It does not flush pending bits or move the underlying writer's cursor.
func (k *Writer) HasPendingBits() bool {
	return k.bitsLeft > 0
}

// Seek seeks to the given position, if the stream is seekable.
func (k *Writer) Seek(offset int64, whence int) (int64, error) {
	err := k.AlignToByte()
	if err != nil {
		return 0, err
	}
	switch w := k.Writer.(type) {
	case io.Seeker:
		n, err := w.Seek(offset, whence)
		if err != nil {
			return 0, fmt.Errorf("Seek: failed to seek stream: %w", err)
		}
		return n, nil
	default:
		return 0, fmt.Errorf("Seek: %w", ErrWriterNotSeekable)
	}
}

// WriteU1 writes a uint8 to the underlying writer.
func (k *Writer) WriteU1(v uint8) error {
	err := k.AlignToByte()
	if err != nil {
		return err
	}
	k.buf[0] = v
	_, err = k.writeNotAligned(k.buf[:1])
	if err != nil {
		return fmt.Errorf("WriteU1: failed to write uint8: %w", err)
	}
	return nil
}

// WriteU2be writes a uint16 in big-endian order to the underlying writer.
func (k *Writer) WriteU2be(v uint16) error {
	err := k.AlignToByte()
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint16(k.buf[:2], v)
	_, err = k.writeNotAligned(k.buf[:2])
	if err != nil {
		return fmt.Errorf("WriteU2be: failed to write uint16: %w", err)
	}
	return nil
}

// WriteU4be writes a uint32 in big-endian order to the underlying writer.
func (k *Writer) WriteU4be(v uint32) error {
	err := k.AlignToByte()
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint32(k.buf[:4], v)
	_, err = k.writeNotAligned(k.buf[:4])
	if err != nil {
		return fmt.Errorf("WriteU4be: failed to write uint32: %w", err)
	}
	return nil
}

// WriteU8be writes a uint64 in big-endian order to the underlying writer.
func (k *Writer) WriteU8be(v uint64) error {
	err := k.AlignToByte()
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint64(k.buf[:8], v)
	_, err = k.writeNotAligned(k.buf[:8])
	if err != nil {
		return fmt.Errorf("WriteU8be: failed to write uint64: %w", err)
	}
	return nil
}

// WriteU2le writes a uint16 in little-endian order to the underlying writer.
func (k *Writer) WriteU2le(v uint16) error {
	err := k.AlignToByte()
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint16(k.buf[:2], v)
	_, err = k.writeNotAligned(k.buf[:2])
	if err != nil {
		return fmt.Errorf("WriteU2le: failed to write uint16: %w", err)
	}
	return nil
}

// WriteU4le writes a uint32 in little-endian order to the underlying writer.
func (k *Writer) WriteU4le(v uint32) error {
	err := k.AlignToByte()
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint32(k.buf[:4], v)
	_, err = k.writeNotAligned(k.buf[:4])
	if err != nil {
		return fmt.Errorf("WriteU4le: failed to write uint32: %w", err)
	}
	return nil
}

// WriteU8le writes a uint64 in little-endian order to the underlying writer.
func (k *Writer) WriteU8le(v uint64) error {
	err := k.AlignToByte()
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint64(k.buf[:8], v)
	_, err = k.writeNotAligned(k.buf[:8])
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
	err := k.AlignToByte()
	if err != nil {
		return err
	}
	_, err = k.writeNotAligned(b)
	if err != nil {
		return fmt.Errorf("WriteBytes: failed to write bytes: %w", err)
	}
	return nil
}

// WriteBytesLimit writes fixed-size data with padding or terminator.
// It returns an error wrapping ErrBufferCapacity if data exceeds size.
// term: terminator byte to write after data (-1 = no terminator).
// padRight: padding byte to fill remaining space (-1 = use term, or 0x00 if none)
func (k *Writer) WriteBytesLimit(data []byte, size int, term int, padRight int) error {
	if len(data) > size {
		return fmt.Errorf(
			"WriteBytesLimit: writing %d bytes, but %d bytes were given: %w",
			size, len(data), ErrBufferCapacity,
		)
	}
	err := k.WriteBytes(data)
	if err != nil {
		return fmt.Errorf("WriteBytesLimit: failed to write bytes: %w", err)
	}
	remaining := size - len(data)
	if remaining <= 0 {
		return nil
	}
	pad := bytesLimitPadding(term, padRight)
	// Write terminator if specified
	if term >= 0 && remaining > 0 {
		err := k.WriteU1(byte(term))
		if err != nil {
			return fmt.Errorf("WriteBytesLimit: failed to write terminator: %w", err)
		}
		remaining--
	}
	// Fill remaining with pad byte
	if remaining > 0 {
		padding := make([]byte, remaining)
		for i := range padding {
			padding[i] = pad
		}
		err := k.WriteBytes(padding)
		if err != nil {
			return fmt.Errorf("WriteBytesLimit: failed to write padding: %w", err)
		}
	}
	return nil
}

func bytesLimitPadding(term, padRight int) byte {
	if padRight >= 0 {
		return byte(padRight)
	}
	if term >= 0 {
		return byte(term)
	}
	return 0
}

// WriteBitsIntBe writes n bits in big-endian bit order. n must be in the
// inclusive range 0..64; otherwise it returns ErrInvalidSizeRequested without
// changing writer state. A zero-width write is a no-op.
func (k *Writer) WriteBitsIntBe(n int, val uint64) error {
	if n < 0 || n > 64 {
		return fmt.Errorf("WriteBitsIntBe(%d): %w", n, ErrInvalidSizeRequested)
	}
	if n == 0 {
		return nil
	}
	err := k.ensureBitsCapacity(n)
	if err != nil {
		return fmt.Errorf("WriteBitsIntBe: %w", err)
	}

	mask := uint64((1 << n) - 1)
	val &= mask

	bitsToWrite := k.bitsLeft + n
	bytesToWrite := bitsToWrite / 8
	newBitsLeft := bitsToWrite % 8

	if bytesToWrite > 0 {
		err = k.rewindReservedBits()
		if err != nil {
			return fmt.Errorf("WriteBitsIntBe: %w", err)
		}
		buf := make([]byte, bytesToWrite)

		mask := uint64((1 << newBitsLeft) - 1) // `newBitsLeft` is in range 0..7
		newBits := val & mask
		val = val>>newBitsLeft | k.bits<<(n-newBitsLeft)
		k.bits = newBits
		k.bitsLeft = newBitsLeft
		k.bitsLe = false

		for i := bytesToWrite - 1; i >= 0; i-- {
			buf[i] = byte(val & 0xFF)
			val >>= 8
		}
		_, err = k.writeNotAligned(buf)
		if err != nil {
			return fmt.Errorf("WriteBitsIntBe: %w", err)
		}
	} else {
		k.bits = k.bits<<n | val
		k.bitsLeft = newBitsLeft
		k.bitsLe = false
	}

	return nil
}

// WriteBitsIntLe writes n bits in little-endian bit order. n must be in the
// inclusive range 0..64; otherwise it returns ErrInvalidSizeRequested without
// changing writer state. A zero-width write is a no-op.
func (k *Writer) WriteBitsIntLe(n int, val uint64) error {
	if n < 0 || n > 64 {
		return fmt.Errorf("WriteBitsIntLe(%d): %w", n, ErrInvalidSizeRequested)
	}
	if n == 0 {
		return nil
	}
	err := k.ensureBitsCapacity(n)
	if err != nil {
		return fmt.Errorf("WriteBitsIntLe: %w", err)
	}

	bitsToWrite := k.bitsLeft + n
	bytesToWrite := bitsToWrite / 8
	oldBitsLeft := k.bitsLeft
	newBitsLeft := bitsToWrite % 8

	if bytesToWrite > 0 {
		err = k.rewindReservedBits()
		if err != nil {
			return fmt.Errorf("WriteBitsIntLe: %w", err)
		}
		buf := make([]byte, bytesToWrite)

		newBits := val >> (n - newBitsLeft)
		val = val<<oldBitsLeft | k.bits
		k.bits = newBits
		k.bitsLeft = newBitsLeft
		k.bitsLe = true

		for i := range bytesToWrite {
			buf[i] = byte(val & 0xFF)
			val >>= 8
		}
		_, err = k.writeNotAligned(buf)
		if err != nil {
			return fmt.Errorf("WriteBitsIntLe: %w", err)
		}
	} else {
		k.bits |= val << oldBitsLeft
		k.bitsLeft = newBitsLeft
		k.bitsLe = true
	}

	var mask uint64 = (1 << k.bitsLeft) - 1 // `bitsLeft` is in range 0..7
	k.bits &= mask

	return nil
}

// AlignToByte flushes any remaining bits, padding with zeros.
func (k *Writer) AlignToByte() error {
	if k.bitsLeft > 0 {
		b := k.bits
		if !k.bitsLe {
			b <<= 8 - k.bitsLeft
		}
		k.buf[0] = byte(b)
		err := k.rewindReservedBits()
		if err != nil {
			return fmt.Errorf("AlignToByte: %w", err)
		}
		n, err := k.writeNotAligned(k.buf[:1])
		if n > 0 {
			k.bitsLeft = 0
			k.bits = 0
		}
		if err != nil {
			return fmt.Errorf("AlignToByte: %w", err)
		}
	}
	return nil
}

// ensureBitsCapacity makes writes to bounded destinations atomic at bit-field
// granularity. Pending bits already account for one logical byte, so only the
// additional bytes needed by this call are checked. On failure, an unreserved
// pending byte is materialized without consuming it; this keeps the physical
// cursor at the same logical position reported by Pos while allowing a later
// retry to rewind and overwrite that byte.
func (k *Writer) ensureBitsCapacity(n int) error {
	bounded, ok := k.Writer.(interface {
		remainingWriteCapacity() (int64, bool)
	})
	if !ok {
		return nil
	}
	remaining, fixed := bounded.remainingWriteCapacity()
	if !fixed {
		return nil
	}

	bitsToWrite := k.bitsLeft + n
	bytesNeeded := 0
	if bitsToWrite > 0 {
		bytesNeeded = (bitsToWrite-1)/8 + 1
	}
	additional := bytesNeeded
	if k.bitsLeft > 0 {
		additional--
		if !k.bitsReserved {
			remaining--
		}
	}
	if int64(additional) <= remaining {
		return nil
	}
	if k.bitsLeft > 0 && !k.bitsReserved {
		err := k.reservePendingBits()
		if err != nil {
			return err
		}
	}
	return ErrBufferCapacity
}

func (k *Writer) reservePendingBits() error {
	b := k.bits
	if !k.bitsLe {
		b <<= 8 - k.bitsLeft
	}
	k.buf[0] = byte(b)
	n, err := k.writeNotAligned(k.buf[:1])
	if n > 0 {
		k.bitsReserved = true
	}
	if err != nil {
		return fmt.Errorf("reserve pending bits: %w", err)
	}
	return nil
}

func (k *Writer) rewindReservedBits() error {
	if !k.bitsReserved {
		return nil
	}
	seeker, ok := k.Writer.(io.Seeker)
	if !ok {
		return fmt.Errorf("rewind reserved bits: %w", ErrWriterNotSeekable)
	}
	_, err := seeker.Seek(-1, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("rewind reserved bits: %w", err)
	}
	k.bitsReserved = false
	return nil
}

// writeNotAligned writes bytes without flushing the pending bit buffer.
func (k *Writer) writeNotAligned(p []byte) (int, error) {
	n, err := k.Writer.Write(p)
	if err == nil && n != len(p) {
		err = io.ErrShortWrite
	}
	return n, err
}
