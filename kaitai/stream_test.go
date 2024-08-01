package kaitai

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestNewStream(t *testing.T) {
	type args struct {
		r io.ReadSeeker
	}
	tests := []struct {
		name string
		args args
		want *Stream
	}{
		{"nil Stream", args{nil}, &Stream{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewStream(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_EOF(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		want    bool
		wantErr bool
	}{
		{"not EOF", NewStream(bytes.NewReader([]byte("test"))), false, false},
		{"EOF", NewStream(bytes.NewReader([]byte(""))), true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.k.EOF()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.EOF() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Stream.EOF() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_Size(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		want    int64
		wantErr bool
	}{
		{"Size", NewStream(bytes.NewReader([]byte("test"))), 4, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.k.Size()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.Size() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Stream.Size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_Pos(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		want    int64
		wantErr bool
	}{
		{"Pos", NewStream(bytes.NewReader([]byte("test"))), 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.k.Pos()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.Pos() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Stream.Pos() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_ReadU1(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   uint8
		wantErr bool
	}{
		{"Read", NewStream(bytes.NewReader([]byte("test"))), 't', false},
		{"empty Read", NewStream(bytes.NewReader([]byte(""))), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadU1()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadU1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadU1() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadU2be(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   uint16
		wantErr bool
	}{
		{"Read", NewStream(bytes.NewReader([]byte{1, 0})), 256, false},
		{"empty Read", NewStream(bytes.NewReader([]byte(""))), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadU2be()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadU2be() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadU2be() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadU4be(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   uint32
		wantErr bool
	}{
		{"Read", NewStream(bytes.NewReader([]byte{0, 0, 0, 1})), 1, false},
		{"empty Read", NewStream(bytes.NewReader([]byte(""))), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadU4be()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadU4be() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadU4be() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadU8be(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   uint64
		wantErr bool
	}{
		{"Read", NewStream(bytes.NewReader([]byte{0, 0, 0, 0, 0, 0, 0, 1})), 1, false},
		{"empty Read", NewStream(bytes.NewReader([]byte(""))), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadU8be()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadU8be() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadU8be() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadU2le(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   uint16
		wantErr bool
	}{
		{"Read", NewStream(bytes.NewReader([]byte{0, 1})), 256, false},
		{"empty Read", NewStream(bytes.NewReader([]byte(""))), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadU2le()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadU2le() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadU2le() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadU4le(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   uint32
		wantErr bool
	}{
		{"Read", NewStream(bytes.NewReader([]byte{1, 0, 0, 0})), 1, false},
		{"empty Read", NewStream(bytes.NewReader([]byte(""))), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadU4le()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadU4le() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadU4le() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadU8le(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   uint64
		wantErr bool
	}{
		{"Read", NewStream(bytes.NewReader([]byte{1, 0, 0, 0, 0, 0, 0, 0})), 1, false},
		{"empty Read", NewStream(bytes.NewReader([]byte(""))), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadU8le()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadU8le() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadU8le() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadS1(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   int8
		wantErr bool
	}{
		{"Read", NewStream(bytes.NewReader([]byte{0xFF})), -1, false},
		{"empty Read", NewStream(bytes.NewReader([]byte(""))), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadS1()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadS1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadS1() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadS2be(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   int16
		wantErr bool
	}{
		{"Read", NewStream(bytes.NewReader([]byte{0xff, 0xff})), -1, false},
		{"empty Read", NewStream(bytes.NewReader([]byte(""))), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadS2be()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadS2be() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadS2be() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadS4be(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   int32
		wantErr bool
	}{
		{"Read", NewStream(bytes.NewReader([]byte{0xff, 0xff, 0xff, 0xff})), -1, false},
		{"empty Read", NewStream(bytes.NewReader([]byte(""))), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadS4be()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadS4be() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadS4be() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadS8be(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   int64
		wantErr bool
	}{
		{"Read", NewStream(bytes.NewReader([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})), -1, false},
		{"empty Read", NewStream(bytes.NewReader([]byte(""))), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadS8be()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadS8be() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadS8be() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadS2le(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   int16
		wantErr bool
	}{
		{"Read", NewStream(bytes.NewReader([]byte{1, 0})), 1, false},
		{"negative Read", NewStream(bytes.NewReader([]byte{0xff, 0xff})), -1, false},
		{"empty Read", NewStream(bytes.NewReader([]byte(""))), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadS2le()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadS2le() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadS2le() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadS4le(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   int32
		wantErr bool
	}{
		{"Read", NewStream(bytes.NewReader([]byte{1, 0, 0, 0})), 1, false},
		{"negative Read", NewStream(bytes.NewReader([]byte{0xff, 0xff, 0xff, 0xff})), -1, false},
		{"empty Read", NewStream(bytes.NewReader([]byte(""))), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadS4le()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadS4le() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadS4le() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadS8le(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   int64
		wantErr bool
	}{
		{"Read", NewStream(bytes.NewReader([]byte{1, 0, 0, 0, 0, 0, 0, 0})), 1, false},
		{"negative Read", NewStream(bytes.NewReader([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})), -1, false},
		{"empty Read", NewStream(bytes.NewReader([]byte(""))), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadS8le()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadS8le() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadS8le() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadF4be(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   float32
		wantErr bool
	}{
		{"ReadF4be", NewStream(bytes.NewReader([]byte{0x3f, 0x80, 0x00, 0x00})), 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadF4be()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadF4be() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadF4be() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadF8be(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   float64
		wantErr bool
	}{
		{"ReadF8be", NewStream(bytes.NewReader([]byte{0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})), 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadF8be()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadF8be() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadF8be() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadF4le(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   float32
		wantErr bool
	}{
		{"ReadF4le", NewStream(bytes.NewReader([]byte{0x00, 0x00, 0x80, 0x3f})), 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadF4le()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadF4le() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadF4le() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadF8le(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		wantV   float64
		wantErr bool
	}{
		{"ReadF8le", NewStream(bytes.NewReader([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0x3f})), 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, err := tt.k.ReadF8le()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadF8le() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotV != tt.wantV {
				t.Errorf("Stream.ReadF8le() = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestStream_ReadBytes(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name    string
		k       *Stream
		args    args
		wantB   []byte
		wantErr bool
	}{
		{"ReadBytes", NewStream(bytes.NewReader([]byte("test"))), args{2}, []byte("te"), false},
		{"negative ReadBytes", NewStream(bytes.NewReader([]byte("test"))), args{-2}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotB, err := tt.k.ReadBytes(tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotB, tt.wantB) {
				t.Errorf("Stream.ReadBytes() = %v, want %v", gotB, tt.wantB)
			}
		})
	}
}

func TestStream_ReadBytesFull(t *testing.T) {
	tests := []struct {
		name    string
		k       *Stream
		want    []byte
		wantErr bool
	}{
		{"ReadBytes", NewStream(bytes.NewReader([]byte("test"))), []byte("test"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.k.ReadBytesFull()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadBytesFull() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Stream.ReadBytesFull() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_ReadBytesPadTerm(t *testing.T) {
	type args struct {
		size        int
		term        byte
		pad         byte
		includeTerm bool
	}
	tests := []struct {
		name    string
		k       *Stream
		args    args
		want    []byte
		wantErr bool
	}{
		{"ReadBytesPadTerm", NewStream(bytes.NewReader([]byte("fooo"))), args{3, 'o', 'x', false}, []byte("f"), false},
		{"ReadBytesPadTerm", NewStream(bytes.NewReader([]byte("fooo"))), args{3, 'x', 'o', false}, []byte("f"), false},
		{"ReadBytesPadTerm", NewStream(bytes.NewReader([]byte("fooo"))), args{3, 'o', 'x', true}, []byte("fo"), false},
		{"ReadBytesPadTerm", NewStream(bytes.NewReader([]byte("fooo"))), args{3, 'x', 'o', true}, []byte("f"), false},
		{"ReadBytesPadTerm", NewStream(bytes.NewReader([]byte("fooo"))), args{-3, 'x', 'o', true}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.k.ReadBytesPadTerm(tt.args.size, tt.args.term, tt.args.pad, tt.args.includeTerm)
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadBytesPadTerm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Stream.ReadBytesPadTerm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_ReadBytesTerm(t *testing.T) {
	type args struct {
		term        byte
		includeTerm bool
		consumeTerm bool
		eosError    bool
	}
	tests := []struct {
		name    string
		k       *Stream
		args    args
		want    []byte
		wantErr bool
	}{
		{"ReadBytesTerm", NewStream(bytes.NewReader([]byte("fooo"))), args{'o', false, false, false}, []byte("f"), false},
		{"ReadBytesTerm", NewStream(bytes.NewReader([]byte("fooo"))), args{'o', true, false, false}, []byte("fo"), false},
		{"ReadBytesTerm", NewStream(bytes.NewReader([]byte("fooo"))), args{'o', false, true, false}, []byte("f"), false},
		{"ReadBytesTerm", NewStream(bytes.NewReader([]byte("fooo"))), args{'o', false, false, true}, []byte("f"), false},
		{"ReadBytesTerm", NewStream(bytes.NewReader([]byte("fooo"))), args{'o', true, true, false}, []byte("fo"), false},
		{"ReadBytesTerm", NewStream(bytes.NewReader([]byte("fooo"))), args{'o', false, true, true}, []byte("f"), false},
		{"ReadBytesTerm", NewStream(bytes.NewReader([]byte("fooo"))), args{'o', true, false, true}, []byte("fo"), false},
		{"ReadBytesTerm", NewStream(bytes.NewReader([]byte("fooo"))), args{'o', true, true, true}, []byte("fo"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.k.ReadBytesTerm(tt.args.term, tt.args.includeTerm, tt.args.consumeTerm, tt.args.eosError)
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadBytesTerm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Stream.ReadBytesTerm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_AlignToByte(t *testing.T) {
	type bitInt struct {
		bits           int
		want           uint64
		isLittleEndian bool
	}
	tests := []struct {
		name   string
		k      *Stream
		fields [2]bitInt // AlignToByte will be called between fields[0] and fields[1]
	}{
		{"ReadBitsIntBe", NewStream(bytes.NewReader([]byte{0b111100_11, 0b0101_0000})), [2]bitInt{
			{6, 0b111100, false},
			// should skip 2 bits (0b11)
			{4, 0b0101, false},
		}},
		{"ReadBitsIntLe", NewStream(bytes.NewReader([]byte{0b11_001111, 0b0000_1010})), [2]bitInt{
			{6, 0b001111, true},
			// should skip 2 bits (0b11)
			{4, 0b1010, true},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, v := range tt.fields {
				if i == 1 {
					tt.k.AlignToByte()
				}
				var gotVal uint64
				var err error
				var methodName string
				if v.isLittleEndian {
					methodName = /**/ "ReadBitsIntLe"
					gotVal, err = tt.k.ReadBitsIntLe(v.bits)
				} else {
					methodName = /**/ "ReadBitsIntBe"
					gotVal, err = tt.k.ReadBitsIntBe(v.bits)
				}
				if err != nil {
					t.Errorf("fields[%v]: Stream.%s(%v) error = %#v", i, methodName, v.bits, err)
					return
				}
				if gotVal != v.want {
					t.Errorf(
						"fields[%v]: Stream.%s(%v) = 0b%0*b, want 0b%0*b", i, methodName, v.bits,
						v.bits, gotVal, v.bits, v.want)
				}
			}
		})
	}
}

func TestStream_ReadBitsInt(t *testing.T) {
	type args struct {
		totalBitsNeeded int
	}
	tests := []struct {
		name    string
		k       *Stream
		args    args
		wantVal uint64
		wantErr bool
	}{
		{"ReadBitsIntBe", NewStream(bytes.NewReader([]byte{0xF0})), args{5}, 0x1E, false},
		{"ReadBitsIntBe", NewStream(bytes.NewReader([]byte{0x12, 0x34, 0x56, 0xFF})), args{24}, 0x123456, false},
		{"ReadBitsIntBe", NewStream(bytes.NewReader([]byte{0xAB, 0xC7})), args{12}, 0xABC, false},
		{"ReadBitsIntBe", NewStream(bytes.NewReader([]byte{1, 2})), args{17}, 0, true},
		{"ReadBitsIntLe", NewStream(bytes.NewReader([]byte{0xF0})), args{5}, 16, false},
		{"ReadBitsIntLe", NewStream(bytes.NewReader([]byte{0x56, 0x34, 0x12, 0xFF})), args{24}, 0x123456, false},
		{"ReadBitsIntLe", NewStream(bytes.NewReader([]byte{0xBC, 0x7A})), args{12}, 0xABC, false},
		{"ReadBitsIntLe", NewStream(bytes.NewReader([]byte{1, 2})), args{17}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotVal uint64
			var err error

			switch tt.name {
			case "ReadBitsIntBe":
				gotVal, err = tt.k.ReadBitsIntBe(tt.args.totalBitsNeeded)
			case "ReadBitsIntLe":
				gotVal, err = tt.k.ReadBitsIntLe(tt.args.totalBitsNeeded)
			default:
				t.Errorf("Unknown test method: %v", tt.name)
				return
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.%s() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if gotVal != tt.wantVal {
				t.Errorf("Stream.%s() = %v, want %v", tt.name, gotVal, tt.wantVal)
			}
		})
	}
}
