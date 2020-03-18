package kaitai

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
)

// APIVersion defines the currently used API version.
const APIVersion = 0x0001

// A Stream represents a sequence of bytes. It encapsulates reading from files
// and memory, stores pointer to its current position, and allows
// reading/writing of various primitives.
type Stream struct {
	io.ReadSeeker
	buf [8]byte

	// Number of bits remaining in buf[0] for sequential calls to ReadBitsInt
	bitsRemaining uint8
}

// NewStream creates and initializes a new Buffer based on r.
func NewStream(r io.ReadSeeker) *Stream {
	return &Stream{ReadSeeker: r}
}

// EOF returns true when the end of the Stream is reached.
func (k *Stream) EOF() (bool, error) {
	curPos, err := k.Pos()
	if err != nil {
		return false, err
	}

	isEOF := false
	_, err = k.ReadU1()
	if err == io.EOF {
		isEOF = true
		err = nil
	}
	if err != nil {
		return false, err
	}

	_, err = k.Seek(curPos, io.SeekStart)
	return isEOF, err
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
		return 0, err
	}
	// Remember position, which is equal to the full length
	fullSize, err := k.Pos()
	if err != nil {
		return fullSize, err
	}
	// Seek back to the current position
	_, err = k.Seek(curPos, io.SeekStart)
	return fullSize, err
}

// Pos returns the current position of the stream.
func (k *Stream) Pos() (int64, error) {
	return k.Seek(0, io.SeekCurrent)
}

// ReadU1 reads 1 byte and returns this as uint8.
func (k *Stream) ReadU1() (v uint8, err error) {
	if _, err = k.Read(k.buf[:1]); err != nil {
		return 0, err
	}
	return k.buf[0], nil
}

// ReadU2be reads 2 bytes in big-endian order and returns those as uint16.
func (k *Stream) ReadU2be() (v uint16, err error) {
	if _, err = k.Read(k.buf[:2]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(k.buf[:2]), nil
}

// ReadU4be reads 4 bytes in big-endian order and returns those as uint32.
func (k *Stream) ReadU4be() (v uint32, err error) {
	if _, err = k.Read(k.buf[:4]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(k.buf[:4]), nil
}

// ReadU8be reads 8 bytes in big-endian order and returns those as uint64.
func (k *Stream) ReadU8be() (v uint64, err error) {
	if _, err = k.Read(k.buf[:8]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(k.buf[:8]), nil
}

// ReadU2le reads 2 bytes in little-endian order and returns those as uint16.
func (k *Stream) ReadU2le() (v uint16, err error) {
	if _, err = k.Read(k.buf[:2]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(k.buf[:2]), nil
}

// ReadU4le reads 4 bytes in little-endian order and returns those as uint32.
func (k *Stream) ReadU4le() (v uint32, err error) {
	if _, err = k.Read(k.buf[:4]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(k.buf[:4]), nil
}

// ReadU8le reads 8 bytes in little-endian order and returns those as uint64.
func (k *Stream) ReadU8le() (v uint64, err error) {
	if _, err = k.Read(k.buf[:8]); err != nil {
		return 0, err
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
		return nil, fmt.Errorf("ReadBytes(%d): negative number of bytes to read", n)
	}

	b = make([]byte, n)
	_, err = io.ReadFull(k, b)
	return b, err
}

// ReadBytesFull reads all remaining bytes and returns those as a byte array.
func (k *Stream) ReadBytesFull() ([]byte, error) {
	return ioutil.ReadAll(k)
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
// set the term bytes is included in the returned byte array. If consumeTerm
// is set the stream continues after the term byte. If eosError is set EOF
// errors result in an error.
func (k *Stream) ReadBytesTerm(term byte, includeTerm, consumeTerm, eosError bool) ([]byte, error) {
	r := bufio.NewReader(k)
	pos, err := k.Pos()
	if err != nil {
		return []byte {}, err
	}
	slice, err := r.ReadBytes(term)

	if err != nil && (err != io.EOF || eosError) {
		return slice, err
	}
	k.Seek(pos + int64(len(slice)), io.SeekStart)
	if !includeTerm {
		slice = slice[:len(slice)-1]
	}
	if !consumeTerm {
		_, err = k.Seek(-1, io.SeekCurrent)
	}
	return slice, err
}

// ReadStrEOS reads the remaining bytes as a string.
func (k *Stream) ReadStrEOS(encoding string) (string, error) {
	buf, err := ioutil.ReadAll(k)

	// Go's string type can contain any bytes.  The Go `range` operator
	// assumes that the encoding is UTF-8 and some standard Go libraries
	// also would like UTF-8.  For now we'll leave any advanced
	// conversions up to the user.
	return string(buf), err
}

// ReadStrByteLimit reads limit number of bytes and returns those as a string.
func (k *Stream) ReadStrByteLimit(limit int, encoding string) (string, error) {
	buf := make([]byte, limit)
	n, err := k.Read(buf)
	return string(buf[:n]), err
}

// AlignToByte discards the remaining bits and starts reading bits at the
// next byte.
func (k *Stream) AlignToByte() {
	k.bitsRemaining = 0
}

// ReadBitsInt reads totalBitsNeeded bits and return those as uint64.
func (k *Stream) ReadBitsInt(totalBitsNeeded uint8) (val uint64, err error) {
	for totalBitsNeeded > 0 {

		// read next byte into buf
		if k.bitsRemaining == 0 {
			// FIXME we could optimize the readBits == 8 case here in the future
			k.bitsRemaining = 8
			_, err = k.Read(k.buf[:1])
			if err != nil {
				return val, err
			}
		}

		// define how many bits should be read
		readBits := totalBitsNeeded % 8
		if readBits == 0 {
			readBits = 8
		}

		// current byte contains all needed bits
		if readBits < k.bitsRemaining {
			val = (val << readBits) | uint64(k.buf[0]>>(k.bitsRemaining-readBits))
			k.bitsRemaining -= readBits
			k.buf[0] &= (1 << k.bitsRemaining) - 1
			// more bytes are needed
		} else {
			readBits = k.bitsRemaining
			k.bitsRemaining = 0
			val = (val << readBits) | uint64(k.buf[0])
		}

		totalBitsNeeded -= readBits
	}
	return val, nil
}

// ReadBitsArray is not implemented yet.
func (k *Stream) ReadBitsArray(n uint) error {
	return nil // TODO: implement
}
