package kaitai

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestNewWriter(t *testing.T) {
	tests := []struct {
		name  string
		want  *Writer
		wantW string
	}{
		{"Test", NewWriter(&bytes.Buffer{}), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if got := NewWriter(w); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWriter() = %v, want %v", got, tt.want)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("NewWriter() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestWriter_WriteU1(t *testing.T) {
	type args struct {
		v uint8
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteU1(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteU1() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteU2be(t *testing.T) {
	type args struct {
		v uint16
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteU2be(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteU2be() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteU4be(t *testing.T) {
	type args struct {
		v uint32
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteU4be(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteU4be() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteU8be(t *testing.T) {
	type args struct {
		v uint64
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteU8be(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteU8be() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteU2le(t *testing.T) {
	type args struct {
		v uint16
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteU2le(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteU2le() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteU4le(t *testing.T) {
	type args struct {
		v uint32
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteU4le(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteU4le() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteU8le(t *testing.T) {
	type args struct {
		v uint64
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteU8le(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteU8le() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteS1(t *testing.T) {
	type args struct {
		v int8
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteS1(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteS1() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteS2be(t *testing.T) {
	type args struct {
		v int16
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteS2be(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteS2be() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteS4be(t *testing.T) {
	type args struct {
		v int32
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteS4be(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteS4be() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteS8be(t *testing.T) {
	type args struct {
		v int64
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteS8be(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteS8be() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteS2le(t *testing.T) {
	type args struct {
		v int16
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteS2le(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteS2le() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteS4le(t *testing.T) {
	type args struct {
		v int32
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteS4le(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteS4le() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteS8le(t *testing.T) {
	type args struct {
		v int64
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteS8le(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteS8le() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteF4be(t *testing.T) {
	type args struct {
		v float32
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteF4be(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteF4be() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteF8be(t *testing.T) {
	type args struct {
		v float64
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteF8be(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteF8be() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteF4le(t *testing.T) {
	type args struct {
		v float32
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteF4le(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteF4le() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_WriteF8le(t *testing.T) {
	type args struct {
		v float64
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteF8le(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteF8le() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_Pos(t *testing.T) {
	t.Run("seekable", func(t *testing.T) {
		sb := NewSeekableBuffer(nil)
		w := NewWriter(sb)
		pos, err := w.Pos()
		if err != nil {
			t.Fatalf("Pos() error = %v", err)
		}
		if pos != 0 {
			t.Errorf("initial Pos() = %d, want 0", pos)
		}
		err = w.WriteU4be(0x12345678)
		if err != nil {
			t.Fatalf("WriteU4be error: %v", err)
		}
		pos, err = w.Pos()
		if err != nil {
			t.Fatalf("Pos() error = %v", err)
		}
		if pos != 4 {
			t.Errorf("Pos() after 4-byte write = %d, want 4", pos)
		}
	})
	t.Run("non-seekable", func(t *testing.T) {
		w := NewWriter(&bytes.Buffer{})
		_, err := w.Pos()
		if err == nil {
			t.Error("Pos() on non-seekable writer should return error")
		}
	})
}

func TestWriter_Seek(t *testing.T) {
	t.Run("seekable", func(t *testing.T) {
		sb := NewSeekableBuffer(nil)
		w := NewWriter(sb)
		err := w.WriteBytes([]byte{0x01, 0x02, 0x03, 0x04})
		if err != nil {
			t.Fatalf("WriteBytes error: %v", err)
		}
		pos, err := w.Seek(1, io.SeekStart)
		if err != nil {
			t.Fatalf("Seek error: %v", err)
		}
		if pos != 1 {
			t.Errorf("Seek returned %d, want 1", pos)
		}
		err = w.WriteU1(0xFF)
		if err != nil {
			t.Fatalf("WriteU1 error: %v", err)
		}
		want := []byte{0x01, 0xFF, 0x03, 0x04}
		if !bytes.Equal(sb.Bytes(), want) {
			t.Errorf("Bytes() = %v, want %v", sb.Bytes(), want)
		}
	})
	t.Run("non-seekable", func(t *testing.T) {
		w := NewWriter(&bytes.Buffer{})
		_, err := w.Seek(0, io.SeekStart)
		if err == nil {
			t.Error("Seek() on non-seekable writer should return error")
		}
	})
}

func TestWriter_WriteBytesLimit(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		size     int
		term     int
		padRight int
		want     []byte
	}{
		{"shorter, zero-padded", []byte{0x01, 0x02}, 4, -1, -1, []byte{0x01, 0x02, 0x00, 0x00}},
		{"shorter, term fills", []byte{0x01}, 4, 0x55, -1, []byte{0x01, 0x55, 0x55, 0x55}},
		{"shorter, padRight only", []byte{0x01, 0x02}, 5, -1, 0xAA, []byte{0x01, 0x02, 0xAA, 0xAA, 0xAA}},
		{"shorter, term + padRight", []byte{0x01}, 4, 0x55, 0xAA, []byte{0x01, 0x55, 0xAA, 0xAA}},
		{"equal length", []byte{0x01, 0x02, 0x03}, 3, -1, -1, []byte{0x01, 0x02, 0x03}},
		{"longer, truncated", []byte{0x01, 0x02, 0x03, 0x04, 0x05}, 3, -1, -1, []byte{0x01, 0x02, 0x03}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			w := NewWriter(buf)
			err := w.WriteBytesLimit(tt.data, tt.size, tt.term, tt.padRight)
			if err != nil {
				t.Fatalf("WriteBytesLimit error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tt.want) {
				t.Errorf("got %v, want %v", buf.Bytes(), tt.want)
			}
		})
	}
}

func TestWriter_WriteBitsIntBe(t *testing.T) {
	t.Run("two nibbles", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		err := w.WriteBitsIntBe(4, 0xA)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		err = w.WriteBitsIntBe(4, 0x5)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if !bytes.Equal(buf.Bytes(), []byte{0xA5}) {
			t.Errorf("got %v, want [A5]", buf.Bytes())
		}
	})

	t.Run("twelve bits then align", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		err := w.WriteBitsIntBe(12, 0xABC)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		err = w.AlignToByte()
		if err != nil {
			t.Fatalf("align err = %v", err)
		}
		if !bytes.Equal(buf.Bytes(), []byte{0xAB, 0xC0}) {
			t.Errorf("got %v, want [AB C0]", buf.Bytes())
		}
	})

	t.Run("mask oversized value", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		// Only the low 4 bits should make it out (0xF, then 0x0 from 0x10's low nibble).
		err := w.WriteBitsIntBe(4, 0xFF)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		err = w.WriteBitsIntBe(4, 0x10)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if !bytes.Equal(buf.Bytes(), []byte{0xF0}) {
			t.Errorf("got %v, want [F0]", buf.Bytes())
		}
	})

	t.Run("round trip via Stream", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		// 5 bits then 5 bits crossing a byte boundary.
		err := w.WriteBitsIntBe(5, 0x15)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		err = w.WriteBitsIntBe(5, 0x0A)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		err = w.AlignToByte()
		if err != nil {
			t.Fatalf("align err = %v", err)
		}
		s := NewStream(bytes.NewReader(buf.Bytes()))
		v1, err := s.ReadBitsIntBe(5)
		if err != nil {
			t.Fatalf("read err: %v", err)
		}
		v2, err := s.ReadBitsIntBe(5)
		if err != nil {
			t.Fatalf("read err: %v", err)
		}
		if v1 != 0x15 || v2 != 0x0A {
			t.Errorf("round-trip: got (%#x, %#x), want (0x15, 0x0A)", v1, v2)
		}
	})

	t.Run("full 64 bits from aligned state", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		val := uint64(0xDEADBEEF8BADF00D)
		err := w.WriteBitsIntBe(64, val)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		want := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x8B, 0xAD, 0xF0, 0x0D}
		if !bytes.Equal(buf.Bytes(), want) {
			t.Errorf("got %v, want %v", buf.Bytes(), want)
		}
	})

	t.Run("64 bits after partial byte", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		err := w.WriteBitsIntBe(4, 0xA)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		err = w.WriteBitsIntBe(64, 0xDEADBEEF8BADF00D)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		err = w.AlignToByte()
		if err != nil {
			t.Fatalf("align err = %v", err)
		}
		want := []byte{0xAD, 0xEA, 0xDB, 0xEE, 0xF8, 0xBA, 0xDF, 0x00, 0xD0}
		if !bytes.Equal(buf.Bytes(), want) {
			t.Errorf("got %v, want %v", buf.Bytes(), want)
		}
	})
}

func TestWriter_WriteBitsIntLe(t *testing.T) {
	t.Run("two nibbles", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		err := w.WriteBitsIntLe(4, 0xA)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		err = w.WriteBitsIntLe(4, 0x5)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if !bytes.Equal(buf.Bytes(), []byte{0x5A}) {
			t.Errorf("got %v, want [5A]", buf.Bytes())
		}
	})

	t.Run("twelve bits then align", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		err := w.WriteBitsIntLe(12, 0xABC)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		err = w.AlignToByte()
		if err != nil {
			t.Fatalf("align err = %v", err)
		}
		if !bytes.Equal(buf.Bytes(), []byte{0xBC, 0x0A}) {
			t.Errorf("got %v, want [BC 0A]", buf.Bytes())
		}
	})

	t.Run("round trip via Stream", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		err := w.WriteBitsIntLe(12, 0xABC)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		err = w.AlignToByte()
		if err != nil {
			t.Fatalf("align err = %v", err)
		}
		s := NewStream(bytes.NewReader(buf.Bytes()))
		v, err := s.ReadBitsIntLe(12)
		if err != nil {
			t.Fatalf("read err: %v", err)
		}
		if v != 0xABC {
			t.Errorf("round-trip got %#x, want 0xABC", v)
		}
	})

	t.Run("64 bits after partial byte", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		err := w.WriteBitsIntLe(4, 0xA)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		err = w.WriteBitsIntLe(64, 0xDEADBEEF8BADF00D)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		err = w.AlignToByte()
		if err != nil {
			t.Fatalf("align err = %v", err)
		}
		s := NewStream(bytes.NewReader(buf.Bytes()))
		v1, err := s.ReadBitsIntLe(4)
		if err != nil {
			t.Fatalf("read v1 err: %v", err)
		}
		v2, err := s.ReadBitsIntLe(64)
		if err != nil {
			t.Fatalf("read v2 err: %v", err)
		}
		if v1 != 0xA {
			t.Errorf("v1 = %#x, want 0xA", v1)
		}
		if v2 != 0xDEADBEEF8BADF00D {
			t.Errorf("v2 = %#x, want 0xDEADBEEF8BADF00D", v2)
		}
	})
}

func TestWriter_AlignToByte(t *testing.T) {
	t.Run("no bits buffered", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		err := w.AlignToByte()
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if buf.Len() != 0 {
			t.Errorf("AlignToByte with no bits wrote %d bytes, want 0", buf.Len())
		}
	})
	t.Run("partial byte padded with zeros on the right", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		err := w.WriteBitsIntBe(3, 0x5)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		err = w.AlignToByte()
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if !bytes.Equal(buf.Bytes(), []byte{0xA0}) {
			t.Errorf("got %v, want [A0]", buf.Bytes())
		}
	})
}

func TestWriter_AlignToByte_LittleEndian(t *testing.T) {
	t.Run("partial byte preserved in low bits", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		err := w.WriteBitsIntLe(3, 0x5)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		err = w.AlignToByte()
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if !bytes.Equal(buf.Bytes(), []byte{0x05}) {
			t.Errorf("got %v, want [05]", buf.Bytes())
		}
	})
}

func TestWriter_WriteBytes(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		k       *Writer
		args    args
		wantErr bool
	}{
		{"Test", NewWriter(&bytes.Buffer{}), args{[]byte("test")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.k.WriteBytes(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteBytes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
