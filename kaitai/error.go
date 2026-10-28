package kaitai

import (
	"errors"
	"fmt"
)

// ErrWriterNotSeekable indicates that a Writer's underlying stream does not
// implement io.Seeker.
var ErrWriterNotSeekable = errors.New("stream is not seekable")

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

// ConsistencyError reports that a generated consistency check found a field
// whose actual value differs from its expected value.
type ConsistencyError struct {
	id       string
	expected any
	actual   any
}

// NewConsistencyError creates a new ConsistencyError instance.
func NewConsistencyError(id string, expected any, actual any) ConsistencyError {
	return ConsistencyError{
		id:       id,
		expected: expected,
		actual:   actual,
	}
}

// ID is a getter of the field or attribute identifier associated with the
// consistency error.
func (e ConsistencyError) ID() string { return e.id }

// Expected is a getter of the expected value associated with the consistency error.
func (e ConsistencyError) Expected() any { return e.expected }

// Actual is a getter of the actual value associated with the consistency error.
func (e ConsistencyError) Actual() any { return e.actual }

func (e ConsistencyError) Error() string {
	return fmt.Sprintf("Check failed: %s, expected: %v, actual: %v", e.id, e.expected, e.actual)
}

// ConsistencyNotCheckedError reports an attempt to write a generated object
// whose consistency has not been checked since its last modification.
type ConsistencyNotCheckedError struct {
	context any
}

// NewConsistencyNotCheckedError creates a new ConsistencyNotCheckedError
// instance. Context may identify the generated object or type and may be nil.
func NewConsistencyNotCheckedError(context any) ConsistencyNotCheckedError {
	return ConsistencyNotCheckedError{context: context}
}

// Context is a getter of the optional generated object or type context.
func (e ConsistencyNotCheckedError) Context() any { return e.context }

func (ConsistencyNotCheckedError) Error() string {
	return "consistency not checked: _check() has not been called since the last modification of the object"
}

type locationInfo struct {
	io      *Stream
	srcPath string
}

func newLocationInfo(io *Stream, srcPath string) locationInfo {
	return locationInfo{
		io,
		srcPath,
	}
}

func (l locationInfo) Io() *Stream { return l.io }

func (l locationInfo) SrcPath() string { return l.srcPath }

func (l locationInfo) msgWithLocation(msg string) string {
	var pos any
	pos, err := l.io.Pos()
	if err != nil {
		pos = "N/A"
	}
	return fmt.Sprintf("%s: at pos %v: %s", l.srcPath, pos, msg)
}

// ValidationFailedError is an interface that all "Validation*Error"s implement.
type ValidationFailedError interface {
	Actual() any
	Io() *Stream
	SrcPath() string
}

func validationFailedMsg(msg string) string {
	return "validation failed: " + msg
}

// ValidationNotEqualError signals validation failure: we required "Actual" value
// to be equal to "Expected", but it turned out that it's not.
type ValidationNotEqualError struct {
	locationInfo

	expected any
	actual   any
}

// NewValidationNotEqualError creates a new ValidationNotEqualError instance.
func NewValidationNotEqualError(
	expected any, actual any, io *Stream, srcPath string) ValidationNotEqualError {
	return ValidationNotEqualError{
		newLocationInfo(io, srcPath),
		expected,
		actual,
	}
}

// Expected is a getter of the expected value associated with the validation error.
func (e ValidationNotEqualError) Expected() any { return e.expected }

// Actual is a getter of the actual value associated with the validation error.
func (e ValidationNotEqualError) Actual() any { return e.actual }

func (e ValidationNotEqualError) Error() string {
	return e.msgWithLocation(
		validationFailedMsg(
			fmt.Sprintf("not equal, expected %v, but got %v", e.expected, e.actual),
		),
	)
}

// ValidationLessThanError signals validation failure: we required "Actual" value
// to be greater than or equal to "Min", but it turned out that it's not.
type ValidationLessThanError struct {
	locationInfo

	min    any
	actual any
}

// NewValidationLessThanError creates a new ValidationLessThanError instance.
func NewValidationLessThanError(
	min any, actual any, io *Stream, srcPath string) ValidationLessThanError {
	return ValidationLessThanError{
		newLocationInfo(io, srcPath),
		min,
		actual,
	}
}

// Min is a getter of the minimum value associated with the validation error.
func (e ValidationLessThanError) Min() any { return e.min }

// Actual is a getter of the actual value associated with the validation error.
func (e ValidationLessThanError) Actual() any { return e.actual }

func (e ValidationLessThanError) Error() string {
	return e.msgWithLocation(
		validationFailedMsg(
			fmt.Sprintf("not in range, min %v, but got %v", e.min, e.actual),
		),
	)
}

// ValidationGreaterThanError signals validation failure: we required "Actual" value
// to be less than or equal to "Max", but it turned out that it's not.
type ValidationGreaterThanError struct {
	locationInfo

	max    any
	actual any
}

// NewValidationGreaterThanError creates a new ValidationGreaterThanError instance.
func NewValidationGreaterThanError(
	max any, actual any, io *Stream, srcPath string) ValidationGreaterThanError {
	return ValidationGreaterThanError{
		newLocationInfo(io, srcPath),
		max,
		actual,
	}
}

// Max is a getter of the maximum value associated with the validation error.
func (e ValidationGreaterThanError) Max() any { return e.max }

// Actual is a getter of the actual value associated with the validation error.
func (e ValidationGreaterThanError) Actual() any { return e.actual }

func (e ValidationGreaterThanError) Error() string {
	return e.msgWithLocation(
		validationFailedMsg(
			fmt.Sprintf("not in range, max %v, but got %v", e.max, e.actual),
		),
	)
}

// ValidationNotAnyOfError signals validation failure: we required "Actual" value
// to be from the list, but it turned out that it's not.
type ValidationNotAnyOfError struct {
	locationInfo

	actual any
}

// NewValidationNotAnyOfError creates a new ValidationNotAnyOfError instance.
func NewValidationNotAnyOfError(actual any, io *Stream, srcPath string) ValidationNotAnyOfError {
	return ValidationNotAnyOfError{
		newLocationInfo(io, srcPath),
		actual,
	}
}

// Actual is a getter of the actual value associated with the validation error.
func (e ValidationNotAnyOfError) Actual() any { return e.actual }

func (e ValidationNotAnyOfError) Error() string {
	return e.msgWithLocation(
		validationFailedMsg(
			fmt.Sprintf("not any of the list, got %v", e.actual),
		),
	)
}

// ValidationNotInEnumError signals validation failure: we required "Actual" value
// to be in the enum, but it turned out that it's not.
type ValidationNotInEnumError struct {
	locationInfo

	actual any
}

// NewValidationNotInEnumError creates a new ValidationNotInEnumError instance.
func NewValidationNotInEnumError(actual any, io *Stream, srcPath string) ValidationNotInEnumError {
	return ValidationNotInEnumError{
		newLocationInfo(io, srcPath),
		actual,
	}
}

// Actual is a getter of the actual value associated with the validation error.
func (e ValidationNotInEnumError) Actual() any { return e.actual }

func (e ValidationNotInEnumError) Error() string {
	return e.msgWithLocation(
		validationFailedMsg(
			fmt.Sprintf("not in the enum, got %v", e.actual),
		),
	)
}

// ValidationExprError signals validation failure: we required "Actual" value
// to match the expression, but it turned out that it doesn't.
type ValidationExprError struct {
	locationInfo

	actual any
}

// NewValidationExprError creates a new ValidationExprError instance.
func NewValidationExprError(actual any, io *Stream, srcPath string) ValidationExprError {
	return ValidationExprError{
		newLocationInfo(io, srcPath),
		actual,
	}
}

// Actual is a getter of the actual value associated with the validation error.
func (e ValidationExprError) Actual() any { return e.actual }

func (e ValidationExprError) Error() string {
	return e.msgWithLocation(
		validationFailedMsg(
			fmt.Sprintf("not matching the expression, got %v", e.actual),
		),
	)
}
