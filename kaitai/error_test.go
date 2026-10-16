package kaitai

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestConsistencyError(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected any
		actual   any
		want     string
	}{
		{"scalar", "header/version", 2, 1, "Check failed: header/version, expected: 2, actual: 1"},
		{"string", "name", "foo", "bar", "Check failed: name, expected: foo, actual: bar"},
		{
			"byte slice", "payload", []byte{0xa0, 0x00}, []byte{0x00, 0xa0},
			"Check failed: payload, expected: [160 0], actual: [0 160]",
		},
		{"nil", "optional", nil, nil, "Check failed: optional, expected: <nil>, actual: <nil>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf("generated check: %w", NewConsistencyError(tt.id, tt.expected, tt.actual))
			var target ConsistencyError
			if !errors.As(err, &target) {
				t.Fatalf("errors.As(%v) did not find ConsistencyError", err)
			}
			if target.ID() != tt.id {
				t.Errorf("ID() = %q, want %q", target.ID(), tt.id)
			}
			if !reflect.DeepEqual(target.Expected(), tt.expected) {
				t.Errorf("Expected() = %#v, want %#v", target.Expected(), tt.expected)
			}
			if !reflect.DeepEqual(target.Actual(), tt.actual) {
				t.Errorf("Actual() = %#v, want %#v", target.Actual(), tt.actual)
			}
			if got := target.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConsistencyNotCheckedError(t *testing.T) {
	type generatedObject struct {
		value int
	}
	object := &generatedObject{value: 42}
	const wantMessage = "consistency not checked: _check() has not been called since the last modification of the object"

	for _, tt := range []struct {
		name    string
		context any
	}{
		{"generated object", object},
		{"no context", nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf("generated write: %w", NewConsistencyNotCheckedError(tt.context))
			var target ConsistencyNotCheckedError
			if !errors.As(err, &target) {
				t.Fatalf("errors.As(%v) did not find ConsistencyNotCheckedError", err)
			}
			if target.Context() != tt.context {
				t.Errorf("Context() = %#v, want %#v", target.Context(), tt.context)
			}
			if got := target.Error(); got != wantMessage {
				t.Errorf("Error() = %q, want %q", got, wantMessage)
			}
		})
	}
}

func TestEndOfStreamError_Error(t *testing.T) {
	tests := []struct {
		name string
		e    EndOfStreamError
		want string
	}{
		{"Test Error", EndOfStreamError{}, "unexpected end of stream"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("EndOfStreamError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUndecidedEndiannessError_Error(t *testing.T) {
	tests := []struct {
		name string
		u    UndecidedEndiannessError
		want string
	}{
		{"Test Error", UndecidedEndiannessError{}, "undecided endianness"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Error(); got != tt.want {
				t.Errorf("UndecidedEndiannessError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_locationInfo_msgWithLocation(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name      string
		l         locationInfo
		args      args
		ioSeekPos int64
		want      string
	}{
		{
			"msg", locationInfo{NewStream(bytes.NewReader([]byte("test"))), "/seq/0"}, args{"something failed"}, 2,
			"/seq/0: at pos 2: something failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.l.io.Seek(tt.ioSeekPos, io.SeekStart)
			if err != nil {
				t.Error(err)
				return
			}
			if got := tt.l.msgWithLocation(tt.args.msg); got != tt.want {
				t.Errorf("locationInfo.msgWithLocation() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidationFailedError_interface(t *testing.T) {
	io := NewStream(bytes.NewReader([]byte("test")))
	actual := -1
	srcPath := "types/header/seq/7"
	tests := []struct {
		name string
		e    any
	}{
		{"ValidationNotEqualError", NewValidationNotEqualError(2, actual, io, srcPath)},
		{"ValidationLessThanError", NewValidationLessThanError(2, actual, io, srcPath)},
		{"ValidationGreaterThanError", NewValidationGreaterThanError(2, actual, io, srcPath)},
		{"ValidationNotAnyOfError", NewValidationNotAnyOfError(actual, io, srcPath)},
		{"ValidationNotInEnumError", NewValidationNotInEnumError(actual, io, srcPath)},
		{"ValidationExprError", NewValidationExprError(actual, io, srcPath)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, ok := tt.e.(ValidationFailedError)
			if ok != true {
				t.Fatalf("Type %T does not implement ValidationFailedError", tt.e)
			}
			if got := e.Actual(); got != actual {
				t.Errorf("%T.Actual() = %v, want %v", e, got, actual)
			}
			if got := e.Io(); got != io {
				t.Errorf("%T.Io() = %p, want %p", e, got, io)
			}
			if got := e.SrcPath(); got != srcPath {
				t.Errorf("%T.SrcPath() = %q, want %q", e, got, srcPath)
			}
		})
	}
}

func TestValidationNotEqualError_Error(t *testing.T) {
	io := NewStream(bytes.NewReader([]byte("test")))
	tests := []struct {
		name string
		e    ValidationNotEqualError
		want string
	}{
		{
			"integers", NewValidationNotEqualError(42, -1, io, "/seq/0"),
			"/seq/0: at pos 0: validation failed: not equal, expected 42, but got -1",
		},
		{
			"byte arrays", NewValidationNotEqualError([]uint8{160, 0}, []uint8{0, 160}, io, "/seq/1"),
			"/seq/1: at pos 0: validation failed: not equal, expected [160 0], but got [0 160]",
		},
		{
			"strings", NewValidationNotEqualError("ba", "ab", io, "/seq/2"),
			"/seq/2: at pos 0: validation failed: not equal, expected ba, but got ab",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("ValidationNotEqualError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidationNotEqualError_Expected(t *testing.T) {
	io := NewStream(bytes.NewReader([]byte("test")))
	tests := []struct {
		name string
		e    ValidationNotEqualError
		want any
	}{
		{"Expected", NewValidationNotEqualError(42, -1, io, "/seq/0"), 42},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Expected(); got != tt.want {
				t.Errorf("ValidationNotEqualError.Expected() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationLessThanError_Error(t *testing.T) {
	io := NewStream(bytes.NewReader([]byte("test")))
	tests := []struct {
		name string
		e    ValidationLessThanError
		want string
	}{
		{
			"integers", NewValidationLessThanError(42, -42, io, "/seq/0"),
			"/seq/0: at pos 0: validation failed: not in range, min 42, but got -42",
		},
		{
			"byte arrays", NewValidationLessThanError([]uint8{160, 0}, []uint8{0, 160}, io, "/seq/1"),
			"/seq/1: at pos 0: validation failed: not in range, min [160 0], but got [0 160]",
		},
		{
			"strings", NewValidationLessThanError("ba", "ab", io, "/seq/2"),
			"/seq/2: at pos 0: validation failed: not in range, min ba, but got ab",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("ValidationLessThanError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}
func TestValidationLessThanError_Min(t *testing.T) {
	io := NewStream(bytes.NewReader([]byte("test")))
	tests := []struct {
		name string
		e    ValidationLessThanError
		want any
	}{
		{"Min", NewValidationLessThanError(42, -1, io, "/seq/0"), 42},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Min(); got != tt.want {
				t.Errorf("ValidationLessThanError.Min() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationGreaterThanError_Error(t *testing.T) {
	io := NewStream(bytes.NewReader([]byte("test")))
	tests := []struct {
		name string
		e    ValidationGreaterThanError
		want string
	}{
		{
			"integers", NewValidationGreaterThanError(-42, 42, io, "/seq/0"),
			"/seq/0: at pos 0: validation failed: not in range, max -42, but got 42",
		},
		{
			"byte arrays", NewValidationGreaterThanError([]uint8{0, 160}, []uint8{160, 0}, io, "/seq/1"),
			"/seq/1: at pos 0: validation failed: not in range, max [0 160], but got [160 0]",
		},
		{
			"strings", NewValidationGreaterThanError("ab", "ba", io, "/seq/2"),
			"/seq/2: at pos 0: validation failed: not in range, max ab, but got ba",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("ValidationGreaterThanError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}
func TestValidationGreaterThanError_Max(t *testing.T) {
	io := NewStream(bytes.NewReader([]byte("test")))
	tests := []struct {
		name string
		e    ValidationGreaterThanError
		want any
	}{
		{"Max", NewValidationGreaterThanError(42, 45, io, "/seq/0"), 42},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Max(); got != tt.want {
				t.Errorf("ValidationGreaterThanError.Max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationNotAnyOfError_Error(t *testing.T) {
	io := NewStream(bytes.NewReader([]byte("test")))
	tests := []struct {
		name string
		e    ValidationNotAnyOfError
		want string
	}{
		{
			"integer", NewValidationNotAnyOfError(-42, io, "/seq/0"),
			"/seq/0: at pos 0: validation failed: not any of the list, got -42",
		},
		{
			"byte array", NewValidationNotAnyOfError([]uint8{0, 160}, io, "/seq/1"),
			"/seq/1: at pos 0: validation failed: not any of the list, got [0 160]",
		},
		{
			"string", NewValidationNotAnyOfError("ab", io, "/seq/2"),
			"/seq/2: at pos 0: validation failed: not any of the list, got ab",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("ValidationNotAnyOfError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidationNotInEnumError_Error(t *testing.T) {
	io := NewStream(bytes.NewReader([]byte("test")))
	tests := []struct {
		name string
		e    ValidationNotInEnumError
		want string
	}{
		{
			"integer", NewValidationNotInEnumError(-42, io, "/seq/0"),
			"/seq/0: at pos 0: validation failed: not in the enum, got -42",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("ValidationNotInEnumError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidationExprError_Error(t *testing.T) {
	io := NewStream(bytes.NewReader([]byte("test")))
	tests := []struct {
		name string
		e    ValidationExprError
		want string
	}{
		{
			"integer", NewValidationExprError(-42, io, "/seq/0"),
			"/seq/0: at pos 0: validation failed: not matching the expression, got -42",
		},
		{
			"byte array", NewValidationExprError([]uint8{0, 160}, io, "/seq/1"),
			"/seq/1: at pos 0: validation failed: not matching the expression, got [0 160]",
		},
		{
			"string", NewValidationExprError("ab", io, "/seq/2"),
			"/seq/2: at pos 0: validation failed: not matching the expression, got ab",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("ValidationExprError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}
