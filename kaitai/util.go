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

	res, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("ProcessZlib: error reading zlib data: %w", err)
	}
	return res, nil
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
