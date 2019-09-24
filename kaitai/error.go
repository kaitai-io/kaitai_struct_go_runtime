package kaitai

// EndOfStreamError is returned when the stream unexpectedly ends.
type EndOfStreamError struct{}

func (EndOfStreamError) Error() string {
	return "unexpected end of stream"
}

// UndecidedEndiannessError occurs when a value has calculated or inherited
// endianness, and the endianness could not be determined.
type UndecidedEndiannessError struct{}

func (UndecidedEndiannessError) Error() string {
	return "undecided endianness"
}

// ValidationNotEqualError is returned when validation fails.
type ValidationNotEqualError struct{}

func (ValidationNotEqualError) Error() string {
	return "validation error"
}
