package kaitai

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
)

// APIVersion defines the currently used API version.
const APIVersion = 0x0001

// ErrInvalidSizeRequested is returned when KaitaiStream methods are called
// with a size argument which does not make sense:
//
// - ReadBytes with negative number of bytes
// - ReadBitsIntBe/Le with more than 8 bytes
var ErrInvalidSizeRequested = errors.New("invalid size requested")

// A Stream represents a sequence of bytes. It encapsulates reading from files
// and memory, stores pointer to its current position, and allows
// reading/writing of various primitives.
type Stream struct {
	io.ReadSeeker
	buf [8]byte

	// Number of bits remaining in "bits" for sequential calls to ReadBitsInt
	bitsLeft int
	bits     uint64
}

// NewStream creates and initializes a new Buffer based on r.
func NewStream(r io.ReadSeeker) *Stream {
	return &Stream{ReadSeeker: r}
}

// EOF returns true when the end of the Stream is reached.
func (k *Stream) EOF() (bool, error) {
	if k.bitsLeft > 0 {
		return false, nil
	}
	curPos, err := k.Pos()
	if err != nil {
		return false, err
	}

	isEOF := false
	_, err = k.ReadU1()
	if errors.Is(err, io.EOF) {
		isEOF = true
		err = nil
	}
	if err != nil {
		return false, err
	}

	_, err = k.Seek(curPos, io.SeekStart)
	if err != nil {
		return false, fmt.Errorf("EOF: error seeking back to current position: %w", err)
	}

	return isEOF, nil
}

// Size returns the number of bytes of the stream.
func (k *Stream) Size() (int64, error) {
	// Go has no internal ReadSeeker function to get current ReadSeeker size,
	// thus we use the following trick.
	// Remember our current position
	curPos, err := k.Pos()
	if err != nil {
		return 0, err
	}
	// Seek to the end of the File object
	_, err = k.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, fmt.Errorf("Size: error seeking to end of the stream: %w", err)
	}
	// Remember position, which is equal to the full length
	fullSize, err := k.Pos()
	if err != nil {
		return fullSize, err
	}
	// Seek back to the current position
	_, err = k.Seek(curPos, io.SeekStart)
	if err != nil {
		return fullSize, fmt.Errorf("Size: error seeking back to current position: %w", err)
	}
	return fullSize, nil
}

// Pos returns the current position of the stream.
func (k *Stream) Pos() (int64, error) {
	pos, err := k.Seek(0, io.SeekCurrent)
	if err != nil {
		return pos, fmt.Errorf("Pos: error getting current position: %w", err)
	}
	return pos, nil
}

// ReadU1 reads 1 byte and returns this as uint8.
func (k *Stream) ReadU1() (v uint8, err error) {
	if _, err = io.ReadFull(k, k.buf[:1]); err != nil {
		return 0, fmt.Errorf("ReadU1: error reading 1 byte: %w", err)
	}
	return k.buf[0], nil
}

// ReadU2be reads 2 bytes in big-endian order and returns those as uint16.
func (k *Stream) ReadU2be() (v uint16, err error) {
	if _, err = io.ReadFull(k, k.buf[:2]); err != nil {
		return 0, fmt.Errorf("ReadU2be: error reading 2 bytes: %w", err)
	}
	return binary.BigEndian.Uint16(k.buf[:2]), nil
}

// ReadU4be reads 4 bytes in big-endian order and returns those as uint32.
func (k *Stream) ReadU4be() (v uint32, err error) {
	if _, err = io.ReadFull(k, k.buf[:4]); err != nil {
		return 0, fmt.Errorf("ReadU4be: error reading 4 bytes: %w", err)
	}
	return binary.BigEndian.Uint32(k.buf[:4]), nil
}

// ReadU8be reads 8 bytes in big-endian order and returns those as uint64.
func (k *Stream) ReadU8be() (v uint64, err error) {
	if _, err = io.ReadFull(k, k.buf[:8]); err != nil {
		return 0, fmt.Errorf("ReadU8be: error reading 8 bytes: %w", err)
	}
	return binary.BigEndian.Uint64(k.buf[:8]), nil
}

// ReadU2le reads 2 bytes in little-endian order and returns those as uint16.
func (k *Stream) ReadU2le() (v uint16, err error) {
	if _, err = io.ReadFull(k, k.buf[:2]); err != nil {
		return 0, fmt.Errorf("ReadU2le: error reading 2 bytes: %w", err)
	}
	return binary.LittleEndian.Uint16(k.buf[:2]), nil
}

// ReadU4le reads 4 bytes in little-endian order and returns those as uint32.
func (k *Stream) ReadU4le() (v uint32, err error) {
	if _, err = io.ReadFull(k, k.buf[:4]); err != nil {
		return 0, fmt.Errorf("ReadU4le: error reading 4 bytes: %w", err)
	}
	return binary.LittleEndian.Uint32(k.buf[:4]), nil
}

// ReadU8le reads 8 bytes in little-endian order and returns those as uint64.
func (k *Stream) ReadU8le() (v uint64, err error) {
	if _, err = io.ReadFull(k, k.buf[:8]); err != nil {
		return 0, fmt.Errorf("ReadU8le: error reading 8 bytes: %w", err)
	}
	return binary.LittleEndian.Uint64(k.buf[:8]), nil
}

// ReadS1 reads 1 byte and returns this as int8.
func (k *Stream) ReadS1() (v int8, err error) {
	vv, err := k.ReadU1()
	return int8(vv), err
}

// ReadS2be reads 2 bytes in big-endian order and returns those as int16.
func (k *Stream) ReadS2be() (v int16, err error) {
	vv, err := k.ReadU2be()
	return int16(vv), err
}

// ReadS4be reads 4 bytes in big-endian order and returns those as int32.
func (k *Stream) ReadS4be() (v int32, err error) {
	vv, err := k.ReadU4be()
	return int32(vv), err
}

// ReadS8be reads 8 bytes in big-endian order and returns those as int64.
func (k *Stream) ReadS8be() (v int64, err error) {
	vv, err := k.ReadU8be()
	return int64(vv), err
}

// ReadS2le reads 2 bytes in little-endian order and returns those as int16.
func (k *Stream) ReadS2le() (v int16, err error) {
	vv, err := k.ReadU2le()
	return int16(vv), err
}

// ReadS4le reads 4 bytes in little-endian order and returns those as int32.
func (k *Stream) ReadS4le() (v int32, err error) {
	vv, err := k.ReadU4le()
	return int32(vv), err
}

// ReadS8le reads 8 bytes in little-endian order and returns those as int64.
func (k *Stream) ReadS8le() (v int64, err error) {
	vv, err := k.ReadU8le()
	return int64(vv), err
}

// ReadF4be reads 4 bytes in big-endian order and returns those as float32.
func (k *Stream) ReadF4be() (v float32, err error) {
	vv, err := k.ReadU4be()
	return math.Float32frombits(vv), err
}

// ReadF8be reads 8 bytes in big-endian order and returns those as float64.
func (k *Stream) ReadF8be() (v float64, err error) {
	vv, err := k.ReadU8be()
	return math.Float64frombits(vv), err
}

// ReadF4le reads 4 bytes in little-endian order and returns those as float32.
func (k *Stream) ReadF4le() (v float32, err error) {
	vv, err := k.ReadU4le()
	return math.Float32frombits(vv), err
}

// ReadF8le reads 8 bytes in little-endian order and returns those as float64.
func (k *Stream) ReadF8le() (v float64, err error) {
	vv, err := k.ReadU8le()
	return math.Float64frombits(vv), err
}

// ReadBytes reads n bytes and returns those as a byte array.
func (k *Stream) ReadBytes(n int) (b []byte, err error) {
	if n < 0 {
		return nil, fmt.Errorf("ReadBytes(%d): %w", n, ErrInvalidSizeRequested)
	}

	b = make([]byte, n)
	_, err = io.ReadFull(k, b)
	if err != nil {
		return nil, fmt.Errorf("ReadBytes: error reading %d bytes: %w", n, err)
	}
	return b, nil
}

// ReadBytesFull reads all remaining bytes and returns those as a byte array.
func (k *Stream) ReadBytesFull() ([]byte, error) {
	res, err := ioutil.ReadAll(k)
	if err != nil {
		return nil, fmt.Errorf("ReadBytesFull: error reading all bytes: %w", err)
	}
	return res, nil
}

// ReadBytesPadTerm reads up to size bytes. pad bytes are discarded. It
// terminates reading, when the term byte occurs. The term byte is included
// in the returned byte array when includeTerm is set.
func (k *Stream) ReadBytesPadTerm(size int, term, pad byte, includeTerm bool) ([]byte, error) {
	bs, err := k.ReadBytes(size)
	if err != nil {
		return nil, err
	}

	bs = bytes.TrimRight(bs, string(pad))

	i := bytes.IndexByte(bs, term)
	if i != -1 {
		if includeTerm {
			bs = bs[:i+1]
		} else {
			bs = bs[:i]
		}
	}

	return bs, nil
}

// ReadBytesTerm reads bytes until the term byte is reached. If includeTerm is
// true, the term byte is included in the returned byte slice. If consumeTerm is
// true, the stream continues after the term byte. If eosError is true, EOF
// errors result in an error.
func (k *Stream) ReadBytesTerm(term byte, includeTerm, consumeTerm, eosError bool) ([]byte, error) {
	r := bufio.NewReader(k)
	pos, err := k.Pos()
	if err != nil {
		return []byte{}, err
	}
	slice, err := r.ReadBytes(term)

	if err != nil {
		// If eosError if false, ignore io.EOF and bail out on any other error
		// If eosError is true, bail out on any error, including io.EOF
		if eosError || !errors.Is(err, io.EOF) {
			return slice, fmt.Errorf("ReadBytesTerm: error reading bytes until term byte: %w", err)
		}
	}
	_, err = k.Seek(pos+int64(len(slice)), io.SeekStart)
	if err != nil {
		return []byte{}, fmt.Errorf("ReadBytesTerm: error seeking past term byte: %w", err)
	}
	if !includeTerm {
		slice = slice[:len(slice)-1]
	}
	if !consumeTerm {
		_, err = k.Seek(-1, io.SeekCurrent)
		if err != nil {
			return slice, fmt.Errorf("ReadBytesTerm: error seeking to term byte: %w", err)
		}
	}
	return slice, nil
}

// ReadBytesTermMulti reads chunks of len(term) bytes until it reaches a chunk
// equal to term. If includeTerm is true, term will be included in the returned
// byte slice. If consumeTerm is true, stream position will be left after the
// term, otherwise a seek will be performed to get the stream position before
// the term. If eosError is true, reaching EOF before the term is found is
// treated as an error, otherwise no error and all bytes until EOF are returned
// in this case.
func (k *Stream) ReadBytesTermMulti(term []byte, includeTerm, consumeTerm, eosError bool) ([]byte, error) {
	unitSize := len(term)
	r := []byte{}
	c := make([]byte, unitSize)
	for {
		n, err := io.ReadFull(k, c)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				if eosError {
					return nil, fmt.Errorf("ReadBytesTermMulti: end of stream reached, but no terminator found")
				}
				r = append(r, c[:n]...)
				return r, nil
			}
			return nil, fmt.Errorf("ReadBytesTermMulti: %w", err)
		}
		if bytes.Equal(c, term) {
			if includeTerm {
				r = append(r, c...)
			}
			if !consumeTerm {
				_, err := k.Seek(-int64(unitSize), io.SeekCurrent)
				if err != nil {
					return nil, fmt.Errorf("ReadBytesTermMulti: error seeking back to terminator: %w", err)
				}
			}
			return r, nil
		}
		r = append(r, c...)
	}
}

// AlignToByte discards the remaining bits and starts reading bits at the
// next byte.
func (k *Stream) AlignToByte() {
	k.bitsLeft = 0
	k.bits = 0
}

// ReadBitsIntBe reads n-bit integer in big-endian byte order and returns it as uint64.
func (k *Stream) ReadBitsIntBe(n int) (res uint64, err error) {
	res = 0

	bitsNeeded := n - k.bitsLeft
	k.bitsLeft = -bitsNeeded & 7 // `-bitsNeeded mod 8`

	if bitsNeeded > 0 {
		// 1 bit  => 1 byte
		// 8 bits => 1 byte
		// 9 bits => 2 bytes
		bytesNeeded := ((bitsNeeded - 1) / 8) + 1 // `ceil(bitsNeeded / 8)`
		if bytesNeeded > 8 {
			return res, fmt.Errorf("ReadBitsIntBe(%d): more than 8 bytes requested: %w", n, ErrInvalidSizeRequested)
		}
		_, err = io.ReadFull(k, k.buf[:bytesNeeded])
		if err != nil {
			return res, fmt.Errorf("ReadBitsIntBe(%d): %w", n, err)
		}
		for i := 0; i < bytesNeeded; i++ {
			res = res<<8 | uint64(k.buf[i])
		}

		newBits := res
		res = res>>k.bitsLeft | k.bits<<bitsNeeded
		k.bits = newBits // will be masked at the end of the function
	} else {
		res = k.bits >> -bitsNeeded // shift unneeded bits out
	}

	var mask uint64 = (1 << k.bitsLeft) - 1 // `bitsLeft` is in range 0..7
	k.bits &= mask

	return res, nil
}

// ReadBitsInt reads n-bit integer in big-endian byte order and returns it as uint64.
//
// Deprecated: Use ReadBitsIntBe instead.
func (k *Stream) ReadBitsInt(n uint8) (res uint64, err error) {
	return k.ReadBitsIntBe(int(n))
}

// ReadBitsIntLe reads n-bit integer in little-endian byte order and returns it as uint64.
func (k *Stream) ReadBitsIntLe(n int) (res uint64, err error) {
	res = 0
	bitsNeeded := n - k.bitsLeft

	if bitsNeeded > 0 {
		// 1 bit  => 1 byte
		// 8 bits => 1 byte
		// 9 bits => 2 bytes
		bytesNeeded := ((bitsNeeded - 1) / 8) + 1 // `ceil(bitsNeeded / 8)`
		if bytesNeeded > 8 {
			return res, fmt.Errorf("ReadBitsIntLe(%d): more than 8 bytes requested: %w", n, ErrInvalidSizeRequested)
		}
		_, err = io.ReadFull(k, k.buf[:bytesNeeded])
		if err != nil {
			return res, fmt.Errorf("ReadBitsIntLe(%d): %w", n, err)
		}
		for i := 0; i < bytesNeeded; i++ {
			res |= uint64(k.buf[i]) << (i * 8)
		}

		newBits := res >> bitsNeeded
		res = res<<k.bitsLeft | k.bits
		k.bits = newBits
	} else {
		res = k.bits
		k.bits >>= n
	}

	k.bitsLeft = -bitsNeeded & 7 // `-bitsNeeded mod 8`

	var mask uint64 = (1 << n) - 1 // unlike some other languages, no problem with this in Go
	res &= mask
	return res, nil
}
