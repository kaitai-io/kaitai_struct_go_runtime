package kaitai

type CustomDecoder interface {
	Encode(src []byte) []byte
}

type CustomProcessor CustomDecoder
