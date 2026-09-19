package kaitai

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"math/bits"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

// ByteEncoderFunc adapts a function to ByteEncoder. The function follows the
// same input ownership and retry contract as ByteEncoder.
type ByteEncoderFunc func([]byte) ([]byte, error)

// Encode applies f to data.
func (f ByteEncoderFunc) Encode(data []byte) ([]byte, error) {
	return f(data)
}

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
		return nil, fmt.Errorf("ProcessZlib: error initializing zlib reader: %w", err)
	}
	defer func() {
		// Close releases the decompressor; any read/checksum error is handled below.
		_ = r.Close()
	}()

	res, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("ProcessZlib: error reading zlib data: %w", err)
	}
	return res, nil
}

// UnprocessZlib compresses data using the RFC 1950 zlib framing. It chooses the
// smaller of the standard compressor output and a final fixed-Huffman literal
// block, which avoids the standard compressor's empty-block overhead for small
// fields. Output is deterministic but is not guaranteed to be byte-identical to
// other zlib implementations or across Go versions.
func UnprocessZlib(in []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	_, err := w.Write(in)
	if err != nil {
		_ = w.Close()
		return nil, fmt.Errorf("UnprocessZlib: error writing zlib data: %w", err)
	}
	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("UnprocessZlib: error flushing zlib stream: %w", err)
	}
	return compactZlib(in, buf.Bytes()), nil
}

// compactZlib constructs a second valid encoding rather than removing bytes
// from DEFLATE blocks. The header and Adler-32 checksum belong to the completed
// standard encoding of the same input and can be retained unchanged.
func compactZlib(in, standard []byte) []byte {
	// Every literal needs at least eight bits. Two framing bytes, four checksum
	// bytes, three block-header bits and seven end-of-block bits add eight bytes.
	if len(in) >= len(standard)-8 {
		return standard
	}
	size := len(in) + 8
	extraBits := 2 // the block header and end-of-block leave two bits in a byte
	for _, b := range in {
		if b >= 144 {
			extraBits++
			if extraBits > 8 {
				size++
				extraBits = 1
			}
		}
		if size >= len(standard) {
			return standard
		}
	}
	encoded := make([]byte, 2, size)
	copy(encoded, standard[:2])
	var pending uint32 = 3 // BFINAL=1, BTYPE=01 (fixed Huffman), least significant bit first
	var count uint = 3
	put := func(code uint16, width uint) {
		pending |= uint32(bits.Reverse16(code)>>(16-width)) << count
		count += width
		for count >= 8 {
			encoded = append(encoded, byte(pending))
			pending >>= 8
			count -= 8
		}
	}
	// RFC 1951 section 3.2.6 defines the literal codes and end-of-block symbol.
	for _, b := range in {
		if b < 144 {
			put(uint16(b)+0x30, 8)
		} else {
			put(uint16(b)-144+0x190, 9)
		}
	}
	put(0, 7) // symbol 256: end of block
	if count > 0 {
		encoded = append(encoded, byte(pending))
	}
	return append(encoded, standard[len(standard)-4:]...)
}

// BytesToStr returns a string decoded by the given decoder.
func BytesToStr(in []byte, decoder *encoding.Decoder) (string, error) {
	i := bytes.NewReader(in)
	o := transform.NewReader(i, decoder)
	d, err := io.ReadAll(o)
	if err != nil {
		return "", fmt.Errorf("BytesToStr: error decoding bytes with %T: %w", decoder.Transformer, err)
	}
	return string(d), nil
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
	termIndex := bytes.IndexByte(s, term)
	if termIndex == -1 {
		return s
	}
	newLen := termIndex
	if includeTerm {
		newLen++
	}
	return s[:newLen]
}

// BytesTerminateMulti terminates the given byte slice using the provided byte
// sequence term, whose first byte must appear at a position that is a multiple
// of len(term). Occurrences at any other positions are ignored. If includeTerm
// is true, term will be included in the returned byte slice.
func BytesTerminateMulti(s, term []byte, includeTerm bool) []byte {
	unitSize := len(term)
	rest := s
	for {
		searchIndex := bytes.Index(rest, term)
		if searchIndex == -1 {
			return s
		}
		mod := searchIndex % unitSize
		if mod == 0 {
			newLen := (len(s) - len(rest)) + searchIndex
			if includeTerm {
				newLen += unitSize
			}
			return s[:newLen]
		}
		rest = rest[searchIndex+(unitSize-mod):]
	}
}

// BytesStripRight strips bytes of a given value off the end of the byte slice.
func BytesStripRight(s []byte, pad byte) []byte {
	n := len(s)
	for n > 0 && s[n-1] == pad {
		n--
	}
	return s[:n]
}
