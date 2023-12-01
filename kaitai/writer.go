package kaitai

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// A Writer encapsulates writing binary data to files and memory.
type Writer struct {
	io.Writer
	n int64

	buf           [8]byte
	bitsLe        bool
	bitsWriteMode bool
	bitsLeft      int
	bits          int64
}

// NewWriter creates and initializes a new Writer using w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{Writer: w}
}

func (k *Writer) AlignToByte() {
	k.bitsLeft = 0
	k.bits = 0
}

// WriteU1 writes a uint8 to the underlying writer.
func (k *Writer) WriteU1(v uint8) error {
	k.buf[0] = v
	_, err := k.Write(k.buf[:1])
	return err
}

// WriteU2be writes a uint16 in big-endian order to the underlying writer.
func (k *Writer) WriteU2be(v uint16) error {
	binary.BigEndian.PutUint16(k.buf[:2], v)
	_, err := k.Write(k.buf[:2])
	return err
}

// WriteU4be writes a uint32 in big-endian order to the underlying writer.
func (k *Writer) WriteU4be(v uint32) error {
	binary.BigEndian.PutUint32(k.buf[:4], v)
	_, err := k.Write(k.buf[:4])
	return err
}

// WriteU8be writes a uint64 in big-endian order to the underlying writer.
func (k *Writer) WriteU8be(v uint64) error {
	binary.BigEndian.PutUint64(k.buf[:8], v)
	_, err := k.Write(k.buf[:8])
	return err
}

// WriteU2le writes a uint16 in little-endian order to the underlying writer.
func (k *Writer) WriteU2le(v uint16) error {
	binary.LittleEndian.PutUint16(k.buf[:2], v)
	_, err := k.Write(k.buf[:2])
	return err
}

// WriteU4le writes a uint32 in little-endian order to the underlying writer.
func (k *Writer) WriteU4le(v uint32) error {
	binary.LittleEndian.PutUint32(k.buf[:4], v)
	_, err := k.Write(k.buf[:4])
	return err
}

// WriteU8le writes a uint64 in little-endian order to the underlying writer.
func (k *Writer) WriteU8le(v uint64) error {
	binary.LittleEndian.PutUint64(k.buf[:8], v)
	_, err := k.Write(k.buf[:8])
	return err
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
	err := k.WriteAlignToByte()
	if err != nil {
		return err
	}
	err = k.writeBytesNotAligned(b)
	return err
}

func (k *Writer) WriteBitsIntBe(n int, val int64) error {
	k.bitsLe = false
	k.bitsWriteMode = true

	if n < 64 {
		var mask int64 = (1 << n) - 1
		val &= mask
	}

	bitsToWrite := k.bitsLeft + n
	bytesNeeded := ((bitsToWrite - 1) / 8) + 1

	pos := k.n
	var err error
	if k.bitsLeft > 0 {
		err = k.ensureBytesLeftToWrite(bytesNeeded-1, pos)
	} else {
		err = k.ensureBytesLeftToWrite(bytesNeeded-0, pos)
	}
	if err != nil {
		return err
	}

	bytesToWrite := bitsToWrite / 8
	k.bitsLeft = bitsToWrite & 7
	if bytesToWrite > 0 {
		var mask int64 = (1 << k.bitsLeft) - 1
		newBits := val & mask

		var num int64
		if n-k.bitsLeft < 64 {
			num = k.bits << (n - k.bitsLeft)
		}
		val = int64(uint64(val)>>uint64(k.bitsLeft)) | num
		k.bits = newBits

		for i := bytesToWrite - 1; i >= 0; i-- {
			k.buf[i] = byte(val & 0xff)
			val = int64(uint64(val) >> 8)
		}
		err = k.writeBytesNotAligned(k.buf[:])
	} else {
		k.bits = k.bits<<n | val
	}

	return err
}

func (k *Writer) WriteBitsIntLe(n int, val int64) error {
	k.bitsLe = true
	k.bitsWriteMode = true

	bitsToWrite := k.bitsLeft + n
	bytesNeeded := ((bitsToWrite - 1) / 8) + 1

	pos := k.n
	var err error
	if k.bitsLeft > 0 {
		err = k.ensureBytesLeftToWrite(bytesNeeded-1, pos)
	} else {
		err = k.ensureBytesLeftToWrite(bytesNeeded-0, pos)
	}
	if err != nil {
		return err
	}

	bytesToWrite := bitsToWrite / 8
	k.bitsLeft = bitsToWrite & 7
	if bytesToWrite > 0 {
		var mask int64 = (1 << k.bitsLeft) - 1
		newBits := val & mask

		var num int64
		if n-k.bitsLeft < 64 {
			num = k.bits << (n - k.bitsLeft)
		}
		val = int64(uint64(val)>>uint64(k.bitsLeft)) | num
		k.bits = newBits

		for i := bytesToWrite - 1; i >= 0; i-- {
			k.buf[i] = byte(val & 0xff)
			val = int64(uint64(val) >> 8)
		}
		err = k.writeBytesNotAligned(k.buf[:])
	} else {
		k.bits = k.bits<<n | val
	}
	if err != nil {
		return err
	}

	var mask int64 = (1 << k.bitsLeft) - 1
	k.bits &= mask
	return nil
}

func (k *Writer) WriteAlignToByte() error {
	var err error
	if k.bitsLeft > 0 {
		b := byte(k.bits)
		if !k.bitsLe {
			b <<= 8 - k.bitsLeft
		}

		k.AlignToByte()
		err = k.writeBytesNotAligned([]byte{b})
	}
	return err
}

func (k *Writer) WriteBytesLimit(buf []byte, size int64, term int8, padByte int8) error {
	bufLen := int64(len(buf))
	k.WriteBytes(buf)

	var err error
	if bufLen < size {
		k.WriteS1(term)
		var padLen int64 = size - bufLen - 1
		for i := int64(0); i < padLen; i++ {
			e := k.WriteS1(padByte)
			if err != nil {
				err = e
				break
			}
		}
	} else {
		if bufLen > size {
			err = fmt.Errorf(
				"Writing %d bytes, but %d bytes were given", size, bufLen,
			)
		}
	}
	return err
}

func (k *Writer) ensureBytesLeftToWrite(n int, pos int64) error {
	bytesLeft := k.n

	if int64(n) > bytesLeft {
		return errors.New(fmt.Sprintf("requested to write %d bytes, but only %d bytesLeft bytes left in the stream", n, bytesLeft))
	}

	return nil
}

func (k *Writer) writeBytesNotAligned(buf []byte) error {
	_, err := k.Write(buf)
	return err
}
