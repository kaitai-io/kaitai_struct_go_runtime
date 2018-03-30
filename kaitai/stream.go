package kaitai

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
	"io/ioutil"
	"math"
	"math/bits"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

const APIVersion = 0x0001

type Stream struct {
	io.ReadSeeker
	buf [8]byte

	// Number of bits remaining in buf[0] for sequential calls to ReadBitsInt
	bitsRemaining uint8
}

func NewStream(r io.ReadSeeker) (s *Stream) {
	s = &Stream{
		ReadSeeker: r,
	}
	return
}

func (k *Stream) EOF() bool {
	// Not sure about this one.  In Go, an io.EOF is returned as
	// an error from a Read() when the EOF is reached.  EOF
	// handling can then be done like this:
	//
	// v, err := k.ReadU1()
	// if err == io.EOF {
	//       // Handle EOF error
	// } else if err != nil {
	//       // Handle all other errors
	// }
	return false
}

func (k *Stream) Pos() (int64, error) {
	return k.Seek(0, io.SeekCurrent)
}

func (k *Stream) ReadU1() (v uint8, err error) {
	if _, err = k.Read(k.buf[:1]); err != nil {
		return 0, err
	}
	return k.buf[0], nil
}

func (k *Stream) ReadU2be() (v uint16, err error) {
	if _, err = k.Read(k.buf[:2]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(k.buf[:2]), nil
}

func (k *Stream) ReadU4be() (v uint32, err error) {
	if _, err = k.Read(k.buf[:4]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(k.buf[:4]), nil
}

func (k *Stream) ReadU8be() (v uint64, err error) {
	if _, err = k.Read(k.buf[:8]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(k.buf[:8]), nil
}

func (k *Stream) ReadU2le() (v uint16, err error) {
	if _, err = k.Read(k.buf[:2]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(k.buf[:2]), nil
}

func (k *Stream) ReadU4le() (v uint32, err error) {
	if _, err = k.Read(k.buf[:4]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(k.buf[:4]), nil
}

func (k *Stream) ReadU8le() (v uint64, err error) {
	if _, err = k.Read(k.buf[:8]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(k.buf[:8]), nil
}

func (k *Stream) ReadS1() (v int8, err error) {
	vv, err := k.ReadU1()
	return int8(vv), err
}

func (k *Stream) ReadS2be() (v int16, err error) {
	vv, err := k.ReadU2be()
	return int16(vv), err
}

func (k *Stream) ReadS4be() (v int32, err error) {
	vv, err := k.ReadU4be()
	return int32(vv), err
}

func (k *Stream) ReadS8be() (v int64, err error) {
	vv, err := k.ReadU8be()
	return int64(vv), err
}

func (k *Stream) ReadS2le() (v int16, err error) {
	vv, err := k.ReadU2le()
	return int16(vv), err
}

func (k *Stream) ReadS4le() (v int32, err error) {
	vv, err := k.ReadU4le()
	return int32(vv), err
}

func (k *Stream) ReadS8le() (v int64, err error) {
	vv, err := k.ReadU8le()
	return int64(vv), err
}

func (k *Stream) ReadF4be() (v float32, err error) {
	vv, err := k.ReadU4be()
	return math.Float32frombits(vv), err
}

func (k *Stream) ReadF8be() (v float64, err error) {
	vv, err := k.ReadU8be()
	return math.Float64frombits(vv), err
}

func (k *Stream) ReadF4le() (v float32, err error) {
	vv, err := k.ReadU4le()
	return math.Float32frombits(vv), err
}

func (k *Stream) ReadF8le() (v float64, err error) {
	vv, err := k.ReadU8le()
	return math.Float64frombits(vv), err
}

func (k *Stream) ReadBytes(n int) (b []byte, err error) {
	b = make([]byte, n)
	_, err = io.ReadFull(k, b)
	return b, err
}

func (k *Stream) ReadBytesFull() ([]byte, error) {
	return ioutil.ReadAll(k)
}

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

func (k *Stream) ReadBytesTerm(term byte, includeTerm, consumeTerm, eosError bool) ([]byte, error) {
	r := bufio.NewReader(k)
	pos, err := k.Pos()
	if err != nil {
		return nil, err
	}
	slice, err := r.ReadBytes(term)
	defer func() {
		newPos := pos + int64(len(slice))
		if consumeTerm {
			newPos += 1
		}
		k.Seek(newPos, io.SeekStart)
	}()
	if err != nil {
		if !eosError && err == io.EOF {
			err = nil
		}
		return slice, err
	}
	if !includeTerm {
		slice = slice[:len(slice)-1]
	}
	if !consumeTerm {
		// bytes.Reader and bufio.Reader have UnreadByte, but Stream does not yet so we Seek back one byte.
		_, err = k.Seek(-1, io.SeekCurrent)
		if err != nil {
			return slice, err
		}
	}
	return slice, nil
}

// Go's string type can contain any bytes.  The Go `range` operator
// assumes that the encoding is UTF-8 and some standard Go libraries
// also would like UTF-8.  For now we'll leave any advanced
// conversions up to the user.
func (k *Stream) ReadStrEOS(encoding string) (string, error) {
	buf, err := ioutil.ReadAll(k)
	return string(buf), err
}

func (k *Stream) ReadStrByteLimit(limit int, encoding string) (string, error) {
	buf := make([]byte, limit)
	n, err := k.Read(buf)
	return string(buf[:n]), err
}

func (k *Stream) AlignToByte() {
	k.bitsRemaining = 0
}

func (k *Stream) ReadBitsInt(n uint8) (val uint64, err error) {
	for n > 0 {
		b := n % 8
		if k.bitsRemaining == 0 {
			// FIXME we could optimize the b == 8 case here in the future
			k.bitsRemaining = 8
			_, err = k.Read(k.buf[:1])
			if err != nil {
				return val, err
			}
		}
		if b < k.bitsRemaining {
			val = (val << b) | uint64(k.buf[0]>>(k.bitsRemaining-b))
			k.bitsRemaining -= b
			k.buf[0] &= (1 << k.bitsRemaining) - 1
		} else {
			b = k.bitsRemaining
			k.bitsRemaining = 0
			val = (val << b) | uint64(k.buf[0])
		}

		n -= b
	}
	return val, nil
}

// FIXME what does this method do?
func (k *Stream) ReadBitsArray(n uint) error {
	return nil
}

func ProcessXOR(data []byte, key []byte) {
	for i := range data {
		data[i] ^= key[i%len(key)]
	}
}

func ProcessRotateLeft(data []byte, amount int) {
	for i := range data {
		data[i] = byte(bits.RotateLeft8(uint8(data[i]), amount))
	}
}

func ProcessRotateRight(data []byte, amount int) {
	ProcessRotateLeft(data, -amount)
}

func ProcessZlib(in []byte) (out []byte, err error) {
	b := bytes.NewReader(in)

	// FIXME zlib.NewReader allocates a bunch of memory.  In the future
	// we could reuse it by using a sync.Pool if this is called in a tight loop.
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(r)
}

func BytesToStr(in []byte, decoder *encoding.Decoder) (out string, err error) {
	i := bytes.NewReader(in)
	o := transform.NewReader(i, decoder)
	d, e := ioutil.ReadAll(o)
	if e != nil {
		return "", e
	}
	return string(d), nil
}
