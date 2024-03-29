package kaitai

// Struct is the common interface guaranteed to be implemented by all types generated
// by Kaitai Struct compiler.
type Struct interface {
	Kaitai_IO() *Stream
}
