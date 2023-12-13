package kaitai

import (
	"bytes"
	"compress/zlib"
	"errors"
	"io"
	"math/bits"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

// ProcessXOR returns data xored with the key.
func ProcessXOR(data []byte, key []byte) []byte {
	out := make([]byte, len(data))
	for i := range data {
		out[i] = data[i] ^ key[i%len(key)]
	}
	return out
}

// ProcessRotateLeft returns the single bytes in data rotated left by
// amount bits.
func ProcessRotateLeft(data []byte, amount int) []byte {
	out := make([]byte, len(data))
	for i := range data {
		out[i] = bits.RotateLeft8(data[i], amount)
	}
	return out
}

// ProcessRotateRight returns the single bytes in data rotated right by
// amount bits.
func ProcessRotateRight(data []byte, amount int) []byte {
	return ProcessRotateLeft(data, -amount)
}

// ProcessZlib decompresses the given bytes as specified in RFC 1950.
func ProcessZlib(in []byte) ([]byte, error) {
	b := bytes.NewReader(in)

	// FIXME zlib.NewReader allocates a bunch of memory.  In the future
	// we could reuse it by using a sync.Pool if this is called in a tight loop.
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(r)
}

func UnprocessZlib(in []byte) ([]byte, error) {
	w := zlib.NewWriter(bytes.NewBuffer(in))
	defer w.Close()

	var out bytes.Buffer
	if _, err := w.Write(in); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

// BytesToStr returns a string decoded by the given decoder.
func BytesToStr(in []byte, decoder *encoding.Decoder) (string, error) {
	i := bytes.NewReader(in)
	o := transform.NewReader(i, decoder)
	d, e := io.ReadAll(o)
	if e != nil {
		return "", e
	}
	return string(d), nil
}

// StrToBytes returns a bytes encoded by the given encoder.
func StrToBytes(in string, encoder *encoding.Encoder) ([]byte, error) {
	i := strings.NewReader(in)
	o := transform.NewReader(i, encoder)
	return io.ReadAll(o)
}

// StringReverse returns the string s in reverse order.
func StringReverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// BytesTerminate terminates the given byte slice using the provided sentinel,
// optionally including the sentinel itself in the terminated byte slice.
func BytesTerminate(s []byte, term byte, includeTerm bool) []byte {
	n, srcLen := 0, len(s)
	for n < srcLen && s[n] != term {
		n++
	}
	if includeTerm && n < srcLen {
		n++
	}
	return s[:n]
}

// BytesStripRight strips bytes of a given value off the end of the byte slice.
func BytesStripRight(s []byte, pad byte) []byte {
	n := len(s)
	for n > 0 && s[n-1] == pad {
		n--
	}
	return s[:n]
}

func ByteArrayCompare(a []byte, b []byte) int {
	return bytes.Compare(a, b)
}

func ByteArrayIndexof(arr []byte, b byte) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == b {
			return i
		}
	}
	return -1
}

type FakeWriter struct {
	io.ReadSeeker
}

func (fw *FakeWriter) Write([]byte) (n int, err error) {
	return 0, errors.New("unsupported write")
}

func NewFakeWriter(reader io.ReadSeeker) io.ReadWriteSeeker {
	return &FakeWriter{reader}
}

type BytesTerminatedType struct {
	Data       []byte
	terminator byte
	include    bool
	consume    bool
	eosError   bool
}

func NewBytesTerminatedType(terminator byte, include bool, consume bool, eosError bool) *BytesTerminatedType {
	return &BytesTerminatedType{
		terminator: terminator,
		include:    include,
		consume:    consume,
		eosError:   eosError,
	}
}

func (b *BytesTerminatedType) Read(in []byte) error {
	b.Data = BytesTerminate(in, b.terminator, b.include)
	return nil
}

func (b BytesTerminatedType) Write() ([]byte, error) {
	if b.include {
		return b.Data, nil
	}
	return append(b.Data, b.terminator), nil
}

func (b BytesTerminatedType) Size() (uint64, error) {
	result := uint64(len(b.Data))
	if b.include {
		return result, nil
	}
	return result + 1, nil
}

func (b BytesTerminatedType) Bytes() []byte {
	return b.Data
}

type String struct {
	Data     string
	encoding encoding.Encoding
}

func (s *String) Read(in []byte) error {
	data, err := BytesToStr(in, s.encoding.NewDecoder())
	if err != nil {
		return err
	}
	s.Data = data
	return nil
}

func (s String) Write() ([]byte, error) {
	data, err := StrToBytes(s.Data, s.encoding.NewEncoder())
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s String) Size() (uint64, error) {
	d, err := s.Write()
	if err != nil {
		return 0, err
	}
	return uint64(len(d)), nil
}

func (s String) String() string {
	return s.Data
}

type StringTerminatedType struct {
	Data       string
	encoding   encoding.Encoding
	terminator byte
	include    bool
	consume    bool
	eosError   bool
}

func NewStringTerminatedType(encoding encoding.Encoding, terminator byte, include bool, consume bool, eosError bool) *StringTerminatedType {
	return &StringTerminatedType{
		encoding:   encoding,
		terminator: terminator,
		include:    include,
		consume:    consume,
		eosError:   eosError,
	}
}

func (s *StringTerminatedType) Read(in []byte) error {
	var err error
	data := BytesTerminate(in, s.terminator, s.include)
	str, err := BytesToStr(data, s.encoding.NewDecoder())
	if err != nil {
		return err
	}
	s.Data = str
	return nil
}

func (s StringTerminatedType) Write() ([]byte, error) {
	data, err := StrToBytes(s.Data, s.encoding.NewEncoder())
	if err != nil {
		return nil, err
	}
	if s.include {
		return data, nil
	}
	return append(data, s.terminator), nil
}

func (s StringTerminatedType) Size() (uint64, error) {
	d, err := s.Write()
	if err != nil {
		return 0, err
	}
	result := uint64(len(d))
	if s.include {
		return result, nil
	}
	return result + 1, nil
}

func (s StringTerminatedType) String() string {
	return s.Data
}
