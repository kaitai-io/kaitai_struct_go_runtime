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
		{"Zero size", NewStream(bytes.NewReader([]byte{})), 0, false},
		{"Small size", NewStream(bytes.NewReader([]byte("test"))), 4, false},
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

type artificialError struct{}

func (e artificialError) Error() string {
	return "artificial error when seeking with io.SeekCurrent after seeking to end"
}

type failingReader struct {
	pos      int64
	mustFail func(fr failingReader, offset int64, whence int) bool
}

func (fr *failingReader) Read(p []byte) (n int, err error) { return 0, nil }
func (fr *failingReader) Seek(offset int64, whence int) (int64, error) {
	if fr.mustFail(*fr, offset, whence) {
		return 0, artificialError{}
	}

	switch {
	case whence == io.SeekCurrent:
		return fr.pos, nil
	case whence == io.SeekStart:
		fr.pos = offset
	default: // whence == io.SeekEnd
		fr.pos = -1
	}

	return fr.pos, nil
}

// No regression test for issue #26
func TestErrorHandlingInStream_Size(t *testing.T) {
	tests := map[string]struct {
		initialPos       int64
		failingCondition func(fr failingReader, offset int64, whence int) bool
		errorCheck       func(err error) bool
		wantFinalPos     int64
	}{
		"fails to get initial position": {
			initialPos: 5,
			failingCondition: func(fr failingReader, offset int64, whence int) bool {
				return whence == io.SeekCurrent && offset == 0
			},
			errorCheck: func(err error) bool {
				_, ok := err.(artificialError)
				return ok
			},
			wantFinalPos: 5,
		},
		"seek to the end fails": {
			initialPos: 5,
			failingCondition: func(fr failingReader, offset int64, whence int) bool {
				return whence == io.SeekEnd
			},
			errorCheck: func(err error) bool {
				_, ok := err.(artificialError)
				return ok
			},
			wantFinalPos: 5,
		},
		"deferred seek to the initial pos fails": {
			initialPos: 5,
			failingCondition: func(fr failingReader, offset int64, whence int) bool {
				return whence == io.SeekStart && fr.pos == -1
			},
			errorCheck: func(err error) bool {
				_, ok := err.(artificialError)
				return !ok
			},
			wantFinalPos: -1,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fr := &failingReader{tt.initialPos, tt.failingCondition}
			s := NewStream(fr)
			_, err := s.Size()

			if err == nil {
				t.Fatal("Expected error, got nothing")
			}

			if !tt.errorCheck(err) {
				t.Fatalf("Expected error of type %T, got one of type %T", artificialError{}, err)
			}

			if fr.pos != tt.wantFinalPos {
				t.Fatalf("Expected position to be %v, got %v", tt.wantFinalPos, fr.pos)
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
		{"Read", NewStream(bytes.NewReader([]byte{1})), 256, false},
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

func TestStream_ReadStrEOS(t *testing.T) {
	type args struct {
		encoding string
	}
	tests := []struct {
		name    string
		k       *Stream
		args    args
		want    string
		wantErr bool
	}{
		{"ReadStrEOS", NewStream(bytes.NewReader([]byte("fooo"))), args{""}, "fooo", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.k.ReadStrEOS(tt.args.encoding)
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadStrEOS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Stream.ReadStrEOS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_ReadStrByteLimit(t *testing.T) {
	type args struct {
		limit    int
		encoding string
	}
	tests := []struct {
		name    string
		k       *Stream
		args    args
		want    string
		wantErr bool
	}{
		{"ReadStrByteLimit", NewStream(bytes.NewReader([]byte("fooo"))), args{2, ""}, "fo", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.k.ReadStrByteLimit(tt.args.limit, tt.args.encoding)
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadStrByteLimit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Stream.ReadStrByteLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_AlignToByte(t *testing.T) {
	tests := []struct {
		name string
		k    *Stream
	}{
		{"AlignToByte", NewStream(bytes.NewReader([]byte{0xFF}))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.k.AlignToByte()
		})
	}
}

func TestStream_ReadBitsIntBe(t *testing.T) {
	type args struct {
		totalBitsNeeded uint8
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, err := tt.k.ReadBitsIntBe(tt.args.totalBitsNeeded)
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadBitsIntBe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVal != tt.wantVal {
				t.Errorf("Stream.ReadBitsIntBe() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestStream_ReadBitsArray(t *testing.T) {
	type args struct {
		n uint
	}
	tests := []struct {
		name    string
		k       *Stream
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.k.ReadBitsArray(tt.args.n); (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadBitsArray() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStream_ReadBitsIntLe(t *testing.T) {
	type args struct {
		n uint8
	}
	tests := []struct {
		name    string
		k       *Stream
		args    args
		wantRes uint64
		wantErr bool
	}{
		{"ReadBitsIntLe", NewStream(bytes.NewReader([]byte{0xF0})), args{5}, 16, false},
		{"ReadBitsIntLe", NewStream(bytes.NewReader([]byte{0x56, 0x34, 0x12, 0xFF})), args{24}, 0x123456, false},
		{"ReadBitsIntLe", NewStream(bytes.NewReader([]byte{0xBC, 0x7A})), args{12}, 0xABC, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := tt.k.ReadBitsIntLe(tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("Stream.ReadBitsIntLe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRes != tt.wantRes {
				t.Errorf("Stream.ReadBitsIntLe() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
