package kaitai

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"
	"math/bits"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

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

func StringReverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
