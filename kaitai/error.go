package kaitai

import "fmt"

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
	var pos interface{}
	pos, err := l.io.Pos()
	if err != nil {
		pos = "N/A"
	}
	return fmt.Sprintf("%s: at pos %v: %s", l.srcPath, pos, msg)
}

// ValidationFailedError is an interface that all "Validation*Error"s implement.
type ValidationFailedError interface {
	Actual() interface{}
	Io() *Stream
	SrcPath() string
}

func validationFailedMsg(msg string) string {
	return "validation failed: " + msg
}

// ValidationNotEqualError signals validation failure: we required "Actual" value
// to be equal to "Expected", but it turned out that it's not.
type ValidationNotEqualError struct {
	expected interface{}
	actual   interface{}
	locationInfo
}

// NewValidationNotEqualError creates a new ValidationNotEqualError instance.
func NewValidationNotEqualError(
	expected interface{}, actual interface{}, io *Stream, srcPath string) ValidationNotEqualError {
	return ValidationNotEqualError{
		expected,
		actual,
		newLocationInfo(io, srcPath),
	}
}

// Expected is a getter of the expected value associated with the validation error.
func (e ValidationNotEqualError) Expected() interface{} { return e.expected }

// Actual is a getter of the actual value associated with the validation error.
func (e ValidationNotEqualError) Actual() interface{} { return e.actual }

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
	min    interface{}
	actual interface{}
	locationInfo
}

// NewValidationLessThanError creates a new ValidationLessThanError instance.
func NewValidationLessThanError(
	min interface{}, actual interface{}, io *Stream, srcPath string) ValidationLessThanError {
	return ValidationLessThanError{
		min,
		actual,
		newLocationInfo(io, srcPath),
	}
}

// Min is a getter of the minimum value associated with the validation error.
func (e ValidationLessThanError) Min() interface{} { return e.min }

// Actual is a getter of the actual value associated with the validation error.
func (e ValidationLessThanError) Actual() interface{} { return e.actual }

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
	max    interface{}
	actual interface{}
	locationInfo
}

// NewValidationGreaterThanError creates a new ValidationGreaterThanError instance.
func NewValidationGreaterThanError(
	max interface{}, actual interface{}, io *Stream, srcPath string) ValidationGreaterThanError {
	return ValidationGreaterThanError{
		max,
		actual,
		newLocationInfo(io, srcPath),
	}
}

// Max is a getter of the maximum value associated with the validation error.
func (e ValidationGreaterThanError) Max() interface{} { return e.max }

// Actual is a getter of the actual value associated with the validation error.
func (e ValidationGreaterThanError) Actual() interface{} { return e.actual }

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
	actual interface{}
	locationInfo
}

// NewValidationNotAnyOfError creates a new ValidationNotAnyOfError instance.
func NewValidationNotAnyOfError(actual interface{}, io *Stream, srcPath string) ValidationNotAnyOfError {
	return ValidationNotAnyOfError{
		actual,
		newLocationInfo(io, srcPath),
	}
}

// Actual is a getter of the actual value associated with the validation error.
func (e ValidationNotAnyOfError) Actual() interface{} { return e.actual }

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
	actual interface{}
	locationInfo
}

// NewValidationNotInEnumError creates a new ValidationNotInEnumError instance.
func NewValidationNotInEnumError(actual interface{}, io *Stream, srcPath string) ValidationNotInEnumError {
	return ValidationNotInEnumError{
		actual,
		newLocationInfo(io, srcPath),
	}
}

// Actual is a getter of the actual value associated with the validation error.
func (e ValidationNotInEnumError) Actual() interface{} { return e.actual }

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
	actual interface{}
	locationInfo
}

// NewValidationExprError creates a new ValidationExprError instance.
func NewValidationExprError(actual interface{}, io *Stream, srcPath string) ValidationExprError {
	return ValidationExprError{
		actual,
		newLocationInfo(io, srcPath),
	}
}

// Actual is a getter of the actual value associated with the validation error.
func (e ValidationExprError) Actual() interface{} { return e.actual }

func (e ValidationExprError) Error() string {
	return e.msgWithLocation(
		validationFailedMsg(
			fmt.Sprintf("not matching the expression, got %v", e.actual),
		),
	)
}
