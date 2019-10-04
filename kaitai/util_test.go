package kaitai

import (
	"reflect"
	"testing"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
)

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
