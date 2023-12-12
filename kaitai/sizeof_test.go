package kaitai

import (
	"testing"
)

func TestSizeOf(t *testing.T) {
	type args struct {
		*Stream
		Uint8   uint8
		Int8    int8
		Bool    bool
		Uint16  uint16
		Int16   int16
		Uint32  uint32
		Int32   int32
		Float32 float32
		Uint64  uint64
		Int64   int64
		Float64 float64
		String  string
		Bytes   []byte
		Slice   []string
		Array   [4]byte
		Args    *args
	}
	result := args{}
	var initLength uint64 = 55

	l, err := SizeOf(result)
	if err != nil {
		t.Fatal(err)
	}
	if l == initLength {
		t.Fatal("wrong len")
	}
	result.String = "123"
	if l == initLength+3 {
		t.Fatal("wrong len")
	}
	result.Bytes = []byte("123")
	if l == initLength+6 {
		t.Fatal("wrong len")
	}
	result.Slice = []string{"1", "2"}
	if l == initLength+8 {
		t.Fatal("wrong len")
	}
	result.Args = &args{}
	if l == 2*initLength+8 {
		t.Fatal("wrong len")
	}
}
