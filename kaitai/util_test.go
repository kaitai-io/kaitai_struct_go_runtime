package kaitai

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
)

func TestByteEncoderFuncWriteBack(t *testing.T) {
	buf := NewSeekableBuffer(nil)
	parent := NewWriter(buf)
	key := []byte{0x80, 0x01}
	child, err := parent.NewChild(3, 3, ByteEncoderFunc(func(data []byte) ([]byte, error) {
		return ProcessXOR(data, key), nil
	}))
	require.NoError(t, err)
	require.NoError(t, child.WriteBytes([]byte{0x00, 0xff, 0x10}))
	require.NoError(t, parent.WriteBackChildWriters())
	require.Equal(t, []byte{0x80, 0xfe, 0x90}, buf.Bytes())
}

func TestByteEncoderFuncProcessorError(t *testing.T) {
	buf := NewSeekableBuffer([]byte{0xaa, 0xbb})
	parent := NewWriter(buf)
	child, err := parent.NewChild(2, 2, ByteEncoderFunc(ProcessZlib))
	require.NoError(t, err)
	require.NoError(t, child.WriteBytes([]byte{0, 0}))
	err = parent.WriteBackChildWriters()
	require.ErrorIs(t, err, zlib.ErrHeader)
	require.ErrorContains(t, err, "WriteBack: encode: ProcessZlib:")
	require.Equal(t, []byte{0xaa, 0xbb}, buf.Bytes())
}

func TestProcessRoundTripsWithoutMutation(t *testing.T) {
	type processorPair struct {
		name   string
		encode ByteEncoderFunc
		decode ByteEncoderFunc
	}
	tests := []processorPair{
		{"zlib", UnprocessZlib, ProcessZlib},
	}
	for _, key := range [][]byte{{0xa5}, {0x00, 0x81, 0xff}} {
		xor := func(data []byte) ([]byte, error) { return ProcessXOR(data, key), nil }
		tests = append(tests, processorPair{fmt.Sprintf("xor_%x", key), xor, xor})
	}
	for _, amount := range []int{-17, -8, -1, 0, 1, 7, 8, 17} {
		left := func(data []byte) ([]byte, error) { return ProcessRotateLeft(data, amount), nil }
		right := func(data []byte) ([]byte, error) { return ProcessRotateRight(data, amount), nil }
		tests = append(tests,
			processorPair{fmt.Sprintf("rotate_left_%d", amount), right, left},
			processorPair{fmt.Sprintf("rotate_right_%d", amount), left, right})
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, data := range [][]byte{
				nil, {}, {0x00, 0x81, 0xff, 0x42}, bytes.Repeat([]byte{0x00, 0x81, 0xff, 0x42}, 65536),
			} {
				t.Run(fmt.Sprintf("size_%d", len(data)), func(t *testing.T) {
					original := bytes.Clone(data)
					encoded, err := tc.encode.Encode(data)
					require.NoError(t, err)
					require.Equal(t, original, data, "encoding modified source")
					encodedOriginal := bytes.Clone(encoded)
					decoded, err := tc.decode.Encode(encoded)
					require.NoError(t, err)
					require.True(t, bytes.Equal(original, decoded), "round-trip mismatch")
					require.Equal(t, encodedOriginal, encoded, "decoding modified source")
					if len(decoded) > 0 {
						decoded[0] ^= 0xff
						require.Equal(t, encodedOriginal, encoded, "decoded output aliases its input")
					}
					if len(encoded) > 0 {
						encoded[0] ^= 0xff
						require.Equal(t, original, data, "encoded output aliases its input")
					}
				})
			}
		})
	}
}

func TestProcessInverseVectors(t *testing.T) {
	data := []byte{0x81, 0x02, 0xff, 0x00, 0x40}
	key := []byte{0x80, 0x01}
	require.Equal(t, []byte{0x01, 0x03, 0x7f, 0x01, 0xc0}, ProcessXOR(data, key))
	require.Equal(t, []byte{0x80, 0x01}, key, "XOR modified key")
	for _, amount := range []int{-7, 1, 9} {
		require.Equal(t, []byte{0x03, 0x04, 0xff, 0x00, 0x80}, ProcessRotateLeft(data, amount))
		require.Equal(t, []byte{0xc0, 0x01, 0xff, 0x00, 0x20}, ProcessRotateRight(data, amount))
	}
	require.Equal(t, data, ProcessRotateLeft(data, 0))
	require.Equal(t, data, ProcessRotateRight(data, 0))
}

func TestProcessZlibWrappedErrors(t *testing.T) {
	// RFC 1950 stream for "goodbye, world", generated independently of UnprocessZlib.
	valid := []byte{0x78, 0x9c, 0x4b, 0xcf, 0xcf, 0x4f, 0x49, 0xaa,
		0x4c, 0xd5, 0x51, 0x28, 0xcf, 0x2f, 0xca, 0x49,
		0x01, 0x00, 0x28, 0xa5, 0x05, 0x5e}
	badChecksum := bytes.Clone(valid)
	badChecksum[len(badChecksum)-1] ^= 1
	for _, tc := range []struct {
		name string
		data []byte
		want error
	}{
		{"empty", nil, io.ErrUnexpectedEOF},
		{"invalid header", []byte{0, 0}, zlib.ErrHeader},
		{"truncated payload", valid[:len(valid)-5], io.ErrUnexpectedEOF},
		{"checksum", badChecksum, zlib.ErrChecksum},
	} {
		t.Run(tc.name, func(t *testing.T) {
			original := bytes.Clone(tc.data)
			got, err := ProcessZlib(tc.data)
			require.ErrorIs(t, err, tc.want)
			require.ErrorContains(t, err, "ProcessZlib:")
			require.Nil(t, got)
			require.Equal(t, original, tc.data)
		})
	}
}

// A custom processor satisfies ByteEncoder directly, without a function adapter.
type xorByteEncoder struct {
	key byte
}

func (p xorByteEncoder) Encode(data []byte) ([]byte, error) {
	return ProcessXOR(data, []byte{p.key}), nil
}

func TestCustomByteEncoderWriteBack(t *testing.T) {
	buf := NewSeekableBuffer(nil)
	parent := NewWriter(buf)
	child, err := parent.NewChild(2, 2, xorByteEncoder{key: 0xff})
	require.NoError(t, err)
	require.NoError(t, child.WriteBytes([]byte{0x00, 0x12}))
	require.NoError(t, parent.WriteBackChildWriters())
	require.Equal(t, []byte{0xff, 0xed}, buf.Bytes())
}

func TestProcessXOR(t *testing.T) {
	type args struct {
		data []byte
		key  []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"Simple XOR", args{[]byte{0x1}, []byte{0x2}}, []byte{0x3}},
		{"Another simple XOR", args{[]byte{0x1}, []byte{0x1}}, []byte{0x0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ProcessXOR(tt.args.data, tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessXOR() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessRotateLeft(t *testing.T) {
	type args struct {
		data   []byte
		amount int
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"Simple Rotate Left", args{[]byte{0x1}, 1}, []byte{0x2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ProcessRotateLeft(tt.args.data, tt.args.amount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessRotateLeft() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessRotateRight(t *testing.T) {
	type args struct {
		data   []byte
		amount int
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"Simple Rotate Right", args{[]byte{0x2}, 1}, []byte{0x1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ProcessRotateRight(tt.args.data, tt.args.amount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessRotateRight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessZlib(t *testing.T) {
	type args struct {
		in []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"Simple Zlib", args{[]byte{0x78, 0x9c, 0x4b, 0xcf, 0xcf, 0x4f, 0x49, 0xaa,
			0x4c, 0xd5, 0x51, 0x28, 0xcf, 0x2f, 0xca, 0x49,
			0x01, 0x00, 0x28, 0xa5, 0x05, 0x5e}}, []byte("goodbye, world"), false},
		{"Wrong input", args{[]byte{0x00}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessZlib(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessZlib() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessZlib() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnprocessZlib(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
	}{
		{"Empty", []byte{}},
		{"Short string", []byte("goodbye, world")},
		{"Binary data", []byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd, 0x00, 0x00, 0x00}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := UnprocessZlib(tt.in)
			if err != nil {
				t.Fatalf("UnprocessZlib() error = %v", err)
			}
			// Round-trip through ProcessZlib to verify validity.
			got, err := ProcessZlib(compressed)
			if err != nil {
				t.Fatalf("ProcessZlib() after compress error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.in) && (len(got) != 0 || len(tt.in) != 0) {
				t.Errorf("round-trip mismatch: got %v, want %v", got, tt.in)
			}
		})
	}
}

func TestBytesToStr(t *testing.T) {
	utf16 := unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM)
	type args struct {
		in      []byte
		decoder *encoding.Decoder
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"UTF-8 decode", args{[]byte("test"), unicode.UTF8.NewDecoder()}, "test", false},
		{"UTF-16 decode", args{[]byte{0xFE, 0xFF}, utf16.NewDecoder()}, "", false},
		{"UTF-16 decode error", args{[]byte("test"), utf16.NewDecoder()}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BytesToStr(tt.args.in, tt.args.decoder)
			if (err != nil) != tt.wantErr {
				t.Errorf("BytesToStr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BytesToStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringReverse(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Simple Reverse", args{"test"}, "tset"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringReverse(tt.args.s); got != tt.want {
				t.Errorf("StringReverse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBytesTerminate(t *testing.T) {
	type args struct {
		s           []byte
		term        byte
		includeTerm bool
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"Simple Terminate", args{[]byte("test   "), ' ', false}, []byte("test")},
		{"Terminate with include", args{[]byte("test   "), ' ', true}, []byte("test ")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BytesTerminate(tt.args.s, tt.args.term, tt.args.includeTerm); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BytesTerminate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBytesStripRight(t *testing.T) {
	type args struct {
		s   []byte
		pad byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"Simple Strip Right", args{[]byte("test   "), ' '}, []byte("test")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BytesStripRight(tt.args.s, tt.args.pad); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BytesStripRight() = %v, want %v", got, tt.want)
			}
		})
	}
}
