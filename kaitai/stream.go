package kaitai

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// APIVersion defines the currently used API version.
const APIVersion = 0x0001

type Stream struct {
	io.ReadWriteSeeker

	buf           [8]byte
	bitsLeft      int
	bits          uint64
	bitsLe        bool
	bitsWriteMode bool
	// childs writeback handler
	childStreams     []*Stream
	writebackHandler *WriteBackHandler
}

type AnyTypeInterface interface {
	Get_io() *Stream
}

func NewStream(rw io.ReadWriteSeeker) *Stream {
	return &Stream{ReadWriteSeeker: rw}
}

func (k *Stream) WriteU1(v uint8) error {
	k.buf[0] = v
	_, err := k.Write(k.buf[:1])
	return err
}

// WriteU2be writes a uint16 in big-endian order to the underlying writer.
func (k *Stream) WriteU2be(v uint16) error {
	binary.BigEndian.PutUint16(k.buf[:2], v)
	_, err := k.Write(k.buf[:2])
	return err
}

// WriteU4be writes a uint32 in big-endian order to the underlying writer.
func (k *Stream) WriteU4be(v uint32) error {
	binary.BigEndian.PutUint32(k.buf[:4], v)
	_, err := k.Write(k.buf[:4])
	return err
}

// WriteU8be writes a uint64 in big-endian order to the underlying writer.
func (k *Stream) WriteU8be(v uint64) error {
	binary.BigEndian.PutUint64(k.buf[:8], v)
	_, err := k.Write(k.buf[:8])
	return err
}

// WriteU2le writes a uint16 in little-endian order to the underlying writer.
func (k *Stream) WriteU2le(v uint16) error {
	binary.LittleEndian.PutUint16(k.buf[:2], v)
	_, err := k.Write(k.buf[:2])
	return err
}

// WriteU4le writes a uint32 in little-endian order to the underlying writer.
func (k *Stream) WriteU4le(v uint32) error {
	binary.LittleEndian.PutUint32(k.buf[:4], v)
	_, err := k.Write(k.buf[:4])
	return err
}

// WriteU8le writes a uint64 in little-endian order to the underlying writer.
func (k *Stream) WriteU8le(v uint64) error {
	binary.LittleEndian.PutUint64(k.buf[:8], v)
	_, err := k.Write(k.buf[:8])
	return err
}

// WriteS1 writes an int8 to the underlying writer.
func (k *Stream) WriteS1(v int8) error {
	return k.WriteU1(uint8(v))
}

// WriteS2be writes an int16 in big-endian order to the underlying writer.
func (k *Stream) WriteS2be(v int16) error {
	return k.WriteU2be(uint16(v))
}

// WriteS4be writes an in32 in big-endian order to the underlying writer.
func (k *Stream) WriteS4be(v int32) error {
	return k.WriteU4be(uint32(v))
}

// WriteS8be writes an int64 in big-endian order to the underlying writer.
func (k *Stream) WriteS8be(v int64) error {
	return k.WriteU8be(uint64(v))
}

// WriteS2le writes an int16 in little-endian order to the underlying writer.
func (k *Stream) WriteS2le(v int16) error {
	return k.WriteU2le(uint16(v))
}

// WriteS4le writes an int32 in little-endian order to the underlying writer.
func (k *Stream) WriteS4le(v int32) error {
	return k.WriteU4le(uint32(v))
}

// WriteS8le writes an int64 in little-endian order to the underlying writer.
func (k *Stream) WriteS8le(v int64) error {
	return k.WriteU8le(uint64(v))
}

// WriteF4be writes a float32 in big-endian order to the underlying writer.
func (k *Stream) WriteF4be(v float32) error {
	return k.WriteU4be(math.Float32bits(v))
}

// WriteF8be writes a float64 in big-endian order to the underlying writer.
func (k *Stream) WriteF8be(v float64) error {
	return k.WriteU8be(math.Float64bits(v))
}

// WriteF4le writes a float32 in little-endian order to the underlying writer.
func (k *Stream) WriteF4le(v float32) error {
	return k.WriteU4le(math.Float32bits(v))
}

// WriteF8le writes a float64 in little-endian order to the underlying writer.
func (k *Stream) WriteF8le(v float64) error {
	return k.WriteU8le(math.Float64bits(v))
}

// WriteBytes writes the byte slice b to the underlying writer.
func (k *Stream) WriteBytes(b []byte) error {
	err := k.WriteAlignToByte()
	if err != nil {
		return err
	}
	err = k.writeBytesNotAligned(b)
	return err
}

func (k *Stream) WriteBitsIntBe(n int, val uint64) error {
	k.bitsLe = false
	k.bitsWriteMode = true

	if n < 64 {
		var mask uint64 = (1 << n) - 1
		val &= mask
	}

	bitsToWrite := k.bitsLeft + n
	bytesNeeded := ((bitsToWrite - 1) / 8) + 1

	pos, err := k.Pos()
	if err != nil {
		return err
	}
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
		var mask uint64 = (1 << k.bitsLeft) - 1
		newBits := val & mask

		var num uint64
		if n-k.bitsLeft < 64 {
			num = k.bits << (n - k.bitsLeft)
		}
		val = (uint64(val) >> uint64(k.bitsLeft)) | num
		k.bits = newBits

		for i := bytesToWrite - 1; i >= 0; i-- {
			k.buf[i] = byte(val & 0xff)
			val = uint64(val) >> 8
		}
		err = k.writeBytesNotAligned(k.buf[:])
	} else {
		k.bits = k.bits<<n | val
	}

	return err
}

func (k *Stream) WriteBitsIntLe(n int, val uint64) error {
	k.bitsLe = true
	k.bitsWriteMode = true

	bitsToWrite := k.bitsLeft + n
	bytesNeeded := ((bitsToWrite - 1) / 8) + 1

	pos, err := k.Pos()
	if err != nil {
		return err
	}
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
		var mask uint64 = (1 << k.bitsLeft) - 1
		newBits := val & mask

		var num uint64
		if n-k.bitsLeft < 64 {
			num = k.bits << (n - k.bitsLeft)
		}
		val = uint64(val)>>uint64(k.bitsLeft) | num
		k.bits = newBits

		for i := bytesToWrite - 1; i >= 0; i-- {
			k.buf[i] = byte(val & 0xff)
			val = uint64(val) >> 8
		}
		err = k.writeBytesNotAligned(k.buf[:])
	} else {
		k.bits = k.bits<<n | val
	}
	if err != nil {
		return err
	}

	var mask uint64 = (1 << k.bitsLeft) - 1
	k.bits &= mask
	return nil
}

// AlignToByte discards the remaining bits and starts reading bits at the
// next byte.
func (k *Stream) AlignToByte() {
	k.bitsLeft = 0
	k.bits = 0
}

func (k *Stream) WriteAlignToByte() error {
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

func (k *Stream) WriteBytesLimit(buf []byte, size int64, term int8, padByte int8) error {
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

func (k *Stream) ensureBytesLeftToWrite(n int, pos int64) error {
	bytesLeft, err := k.Pos()
	if err != nil {
		return err
	}

	if int64(n) > bytesLeft {
		return errors.New(fmt.Sprintf("requested to write %d bytes, but only %d bytesLeft bytes left in the stream", n, bytesLeft))
	}

	return nil
}

func (k *Stream) writeBytesNotAligned(buf []byte) error {
	_, err := k.Write(buf)
	return err
}

func (k *Stream) WriteBackChildStreams() error {
	return k.writeBackChildStreams(nil)
}

func (k *Stream) writeBackChildStreams(parent *Stream) error {
	pos, err := k.Pos()
	if err != nil {
		return err
	}

	for _, child := range k.childStreams {
		err = child.writeBackChildStreams(k)
		if err != nil {
			return err
		}
	}
	k.childStreams = k.childStreams[:0]
	_, err = k.Seek(pos, io.SeekStart)
	if err != nil {
		return err
	}
	if parent != nil {
		err = k.writeback(parent)
	}

	return err
}

func (k *Stream) writeback(parent *Stream) error {
	return k.writebackHandler.writeBack(parent)
}

func (k *Stream) SetWriteBackHandler(handler *WriteBackHandler) {
	k.writebackHandler = handler
}

func (k *Stream) AddChildStream(child *Stream) {
	if child.ReadWriteSeeker == nil {
		child.ReadWriteSeeker = k.ReadWriteSeeker
	}
	k.childStreams = append(k.childStreams, child)
}

// read part

// ReadU1 reads 1 byte and returns this as uint8.
func (k *Stream) ReadU1() (v uint8, err error) {
	n, err := k.Read(k.buf[:1])
	if err != nil {
		return 0, err
	}
	if n != 1 {
		leftToRead := 1 - n
		leftBuf := k.buf[leftToRead:1]
		leftBuf = leftBuf[0:0]
		_, err = k.Read(k.buf[leftToRead:1])
		if err != nil {
			return 0, err
		}
	}
	return k.buf[0], nil
}

// ReadU2be reads 2 bytes in big-endian order and returns those as uint16.
func (k *Stream) ReadU2be() (v uint16, err error) {
	n, err := k.Read(k.buf[:2])
	if err != nil {
		return 0, err
	}
	if n != 2 {
		leftToRead := 2 - n
		leftBuf := k.buf[leftToRead:2]
		leftBuf = leftBuf[0:0]
		_, err = k.Read(k.buf[leftToRead:2])
		if err != nil {
			return 0, err
		}
	}
	return binary.BigEndian.Uint16(k.buf[:2]), nil
}

// ReadU4be reads 4 bytes in big-endian order and returns those as uint32.
func (k *Stream) ReadU4be() (v uint32, err error) {
	n, err := k.Read(k.buf[:4])
	if err != nil {
		return 0, err
	}
	if n != 4 {
		leftToRead := 4 - n
		leftBuf := k.buf[leftToRead:4]
		leftBuf = leftBuf[0:0]
		_, err = k.Read(k.buf[leftToRead:4])
		if err != nil {
			return 0, err
		}
	}
	return binary.BigEndian.Uint32(k.buf[:4]), nil
}

// ReadU8be reads 8 bytes in big-endian order and returns those as uint64.
func (k *Stream) ReadU8be() (v uint64, err error) {
	n, err := k.Read(k.buf[:8])
	if err != nil {
		return 0, err
	}
	if n != 8 {
		leftToRead := 8 - n
		leftBuf := k.buf[leftToRead:8]
		leftBuf = leftBuf[0:0]
		_, err = k.Read(k.buf[leftToRead:8])
		if err != nil {
			return 0, err
		}
	}
	return binary.BigEndian.Uint64(k.buf[:8]), nil
}

// ReadU2le reads 2 bytes in little-endian order and returns those as uint16.
func (k *Stream) ReadU2le() (v uint16, err error) {
	n, err := k.Read(k.buf[:2])
	if err != nil {
		return 0, err
	}
	if n != 2 {
		leftToRead := 2 - n
		leftBuf := k.buf[leftToRead:2]
		leftBuf = leftBuf[0:0]
		_, err = k.Read(k.buf[leftToRead:2])
		if err != nil {
			return 0, err
		}
	}
	return binary.LittleEndian.Uint16(k.buf[:2]), nil
}

// ReadU4le reads 4 bytes in little-endian order and returns those as uint32.
func (k *Stream) ReadU4le() (v uint32, err error) {
	n, err := k.Read(k.buf[:4])
	if err != nil {
		return 0, err
	}
	if n != 4 {
		leftToRead := 4 - n
		leftBuf := k.buf[leftToRead:4]
		leftBuf = leftBuf[0:0]
		_, err = k.Read(k.buf[leftToRead:4])
		if err != nil {
			return 0, err
		}
	}
	return binary.LittleEndian.Uint32(k.buf[:4]), nil
}

// ReadU8le reads 8 bytes in little-endian order and returns those as uint64.
func (k *Stream) ReadU8le() (v uint64, err error) {
	n, err := k.Read(k.buf[:8])
	if err != nil {
		return 0, err
	}
	if n != 8 {
		leftToRead := 8 - n
		leftBuf := k.buf[leftToRead:8]
		leftBuf = leftBuf[0:0]
		_, err = k.Read(k.buf[leftToRead:8])
		if err != nil {
			return 0, err
		}
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
	return io.ReadAll(k)
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
		return []byte{}, err
	}
	slice, err := r.ReadBytes(term)

	if err != nil && (err != io.EOF || eosError) {
		return slice, err
	}
	_, err = k.Seek(pos+int64(len(slice)), io.SeekStart)
	if err != nil {
		return []byte{}, err
	}
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
	buf, err := io.ReadAll(k)

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
			return res, fmt.Errorf("ReadBitsIntBe(%d): more than 8 bytes requested", n)
		}
		_, err = k.Read(k.buf[:bytesNeeded])
		if err != nil {
			return res, err
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

	return res, err
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
			return res, fmt.Errorf("ReadBitsIntLe(%d): more than 8 bytes requested", n)
		}
		_, err = k.Read(k.buf[:bytesNeeded])
		if err != nil {
			return res, err
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
	return res, err
}

// ReadBitsArray is not implemented yet.
func (k *Stream) ReadBitsArray(n uint) error {
	// TODO: implement, and did not find in https://github1s.com/kaitai-io/kaitai_struct_java_runtime, maybe this is a historical problem?
	return nil
}

// Pos returns the current position of the stream.
func (k *Stream) Pos() (int64, error) {
	return k.Seek(0, io.SeekCurrent)
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

func (k *Stream) ToByteArray() ([]byte, error) {
	pos, err := k.Pos()
	if err != nil {
		return nil, err
	}
	_, err = k.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	b, err := k.ReadBytesFull()
	if err != nil {
		return nil, err
	}
	_, err = k.Seek(pos, io.SeekStart)
	return b, err
}
