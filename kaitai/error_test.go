package kaitai

import "testing"

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

func TestValidationNotEqualError_Error(t *testing.T) {
	tests := []struct {
		name string
		v    ValidationNotEqualError
		want string
	}{
		{"Test Error", ValidationNotEqualError{}, "validation error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.Error(); got != tt.want {
				t.Errorf("ValidationNotEqualError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
