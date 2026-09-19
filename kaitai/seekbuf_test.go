package kaitai

import (
	"bytes"
	"errors"
	"io"
	"math"
	"testing"
)

func TestNewSeekableBuffer(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	sb := NewSeekableBuffer(data)
	if !bytes.Equal(sb.Bytes(), data) {
		t.Errorf("Bytes() = %v, want %v", sb.Bytes(), data)
	}
	if sb.Len() != len(data) {
		t.Errorf("Len() = %d, want %d", sb.Len(), len(data))
	}
	pos, err := sb.Seek(0, io.SeekCurrent)
	if err != nil {
		t.Fatalf("Seek returned error: %v", err)
	}
	if pos != 0 {
		t.Errorf("initial position = %d, want 0", pos)
	}
}

func TestFixedSeekableBufferInitialization(t *testing.T) {
	for _, size := range []int{0, 1, 16} {
		sb := NewFixedSeekableBuffer(size)
		if sb.Len() != size || !bytes.Equal(sb.Bytes(), make([]byte, size)) {
			t.Fatalf("size %d: initial bytes = %v, length = %d", size, sb.Bytes(), sb.Len())
		}
		pos, err := sb.Seek(0, io.SeekCurrent)
		if err != nil || pos != 0 {
			t.Fatalf("size %d: initial position = %d, %v", size, pos, err)
		}
	}
}

func TestFixedSeekableBufferWrite(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		pos      int64
		data     []byte
		want     []byte
		wantPos  int64
		overflow bool
	}{
		{"overwrite", 4, 1, []byte{9, 8}, []byte{1, 9, 8, 4}, 3, false},
		{"exact end", 4, 2, []byte{9, 8}, []byte{1, 2, 9, 8}, 4, false},
		{"whole buffer", 4, 0, []byte{9, 8, 7, 6}, []byte{9, 8, 7, 6}, 4, false},
		{"partial overflow", 4, 2, []byte{9, 8, 7}, []byte{1, 2, 3, 4}, 2, true},
		{"at end", 4, 4, []byte{9}, []byte{1, 2, 3, 4}, 4, true},
		{"empty at end", 4, 4, nil, []byte{1, 2, 3, 4}, 4, false},
		{"zero capacity", 0, 0, []byte{9}, []byte{}, 0, true},
		{"empty zero capacity", 0, 0, nil, []byte{}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewFixedSeekableBuffer(tt.size)
			copy(sb.Bytes(), []byte{1, 2, 3, 4})
			_, err := sb.Seek(tt.pos, io.SeekStart)
			if err != nil {
				t.Fatal(err)
			}
			n, err := sb.Write(tt.data)
			wantN := len(tt.data)
			var wantErr error
			if tt.overflow {
				wantN = 0
				wantErr = ErrBufferCapacity
			}
			if n != wantN || !errors.Is(err, wantErr) {
				t.Fatalf("Write = (%d, %v), want (%d, %v)", n, err, wantN, wantErr)
			}
			if !bytes.Equal(sb.Bytes(), tt.want) || sb.Len() != tt.size {
				t.Errorf("bytes = %v, length = %d; want %v, %d", sb.Bytes(), sb.Len(), tt.want, tt.size)
			}
			pos, err := sb.Seek(0, io.SeekCurrent)
			if err != nil || pos != tt.wantPos {
				t.Errorf("position = (%d, %v), want %d", pos, err, tt.wantPos)
			}
		})
	}
}

func TestFixedSeekableBufferSeek(t *testing.T) {
	tests := []struct {
		name    string
		offset  int64
		whence  int
		want    int64
		wantErr bool
	}{
		{"start", 1, io.SeekStart, 1, false},
		{"current forward", 1, io.SeekCurrent, 3, false},
		{"current backward", -1, io.SeekCurrent, 1, false},
		{"end backward", -1, io.SeekEnd, 3, false},
		{"exact end", 0, io.SeekEnd, 4, false},
		{"exact start", -4, io.SeekEnd, 0, false},
		{"negative start", -1, io.SeekStart, 2, true},
		{"negative current", -3, io.SeekCurrent, 2, true},
		{"negative end", -5, io.SeekEnd, 2, true},
		{"past start", 5, io.SeekStart, 2, true},
		{"past current", 3, io.SeekCurrent, 2, true},
		{"past end", 1, io.SeekEnd, 2, true},
		{"invalid whence", 0, 99, 2, true},
		{"current integer overflow", math.MaxInt64, io.SeekCurrent, 2, true},
		{"end integer overflow", math.MaxInt64, io.SeekEnd, 2, true},
		{"current minimum offset", math.MinInt64, io.SeekCurrent, 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewFixedSeekableBuffer(4)
			_, err := sb.Seek(2, io.SeekStart)
			if err != nil {
				t.Fatal(err)
			}
			pos, err := sb.Seek(tt.offset, tt.whence)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Seek = (%d, %v), wantErr %v", pos, err, tt.wantErr)
			}
			if err == nil && pos != tt.want {
				t.Errorf("Seek position = %d, want %d", pos, tt.want)
			}
			pos, err = sb.Seek(0, io.SeekCurrent)
			if err != nil || pos != tt.want {
				t.Errorf("stored position = (%d, %v), want %d", pos, err, tt.want)
			}
		})
	}
}

func TestFixedSeekableBufferRoundTrip(t *testing.T) {
	sb := NewFixedSeekableBuffer(4)
	var rw io.ReadWriteSeeker = sb
	n, err := rw.Write([]byte{1, 2, 3, 4})
	if err != nil || n != 4 {
		t.Fatalf("Write = (%d, %v)", n, err)
	}
	_, err = rw.Seek(-3, io.SeekEnd)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(sb.Bytes(), []byte{1, 2, 3, 4}) {
		t.Fatalf("Bytes = %v", sb.Bytes())
	}
	// Bytes must neither consume data nor reset the read position.
	got := make([]byte, 5)
	n, err = rw.Read(got)
	if err != nil || n != 3 || !bytes.Equal(got[:n], []byte{2, 3, 4}) {
		t.Fatalf("Read = (%d, %v, %v)", n, err, got)
	}
	n, err = rw.Read(got)
	if n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("Read at end = (%d, %v), want (0, EOF)", n, err)
	}
}

func TestNewSeekableBufferSize(t *testing.T) {
	sb := NewSeekableBufferSize(16)
	if sb.Len() != 16 {
		t.Errorf("Len() = %d, want 16", sb.Len())
	}
	for i, b := range sb.Bytes() {
		if b != 0 {
			t.Errorf("Bytes()[%d] = %#x, want 0", i, b)
		}
	}
}

func TestSeekableBufferSizeStillGrows(t *testing.T) {
	sb := NewSeekableBufferSize(2)
	_, err := sb.Seek(3, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}
	n, err := sb.Write([]byte{9})
	if err != nil || n != 1 {
		t.Fatalf("Write = (%d, %v), want (1, nil)", n, err)
	}
	if !bytes.Equal(sb.Bytes(), []byte{0, 0, 0, 9}) {
		t.Errorf("growable buffer bytes = %v, want [0 0 0 9]", sb.Bytes())
	}
}

func TestSeekableBuffer_Read(t *testing.T) {
	sb := NewSeekableBuffer([]byte{0x10, 0x20, 0x30, 0x40})

	out := make([]byte, 2)
	n, err := sb.Read(out)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if n != 2 || !bytes.Equal(out, []byte{0x10, 0x20}) {
		t.Errorf("first Read = (%d, %v), want (2, [10 20])", n, out)
	}

	out = make([]byte, 8)
	n, err = sb.Read(out)
	if err != nil {
		t.Fatalf("second Read error: %v", err)
	}
	if n != 2 || !bytes.Equal(out[:n], []byte{0x30, 0x40}) {
		t.Errorf("second Read = (%d, %v), want (2, [30 40])", n, out[:n])
	}

	n, err = sb.Read(out)
	if !errors.Is(err, io.EOF) {
		t.Errorf("expected io.EOF at end, got n=%d err=%v", n, err)
	}
}

func TestSeekableBuffer_Write(t *testing.T) {
	sb := NewSeekableBuffer(nil)
	n, err := sb.Write([]byte{0xAA, 0xBB})
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != 2 {
		t.Errorf("Write n = %d, want 2", n)
	}
	if !bytes.Equal(sb.Bytes(), []byte{0xAA, 0xBB}) {
		t.Errorf("Bytes() = %v, want [AA BB]", sb.Bytes())
	}

	_, err = sb.Write([]byte{0xCC, 0xDD, 0xEE})
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if !bytes.Equal(sb.Bytes(), []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE}) {
		t.Errorf("Bytes() = %v, want [AA BB CC DD EE]", sb.Bytes())
	}
}

func TestSeekableBuffer_WriteOverwrite(t *testing.T) {
	sb := NewSeekableBuffer([]byte{0x01, 0x02, 0x03, 0x04})
	_, err := sb.Seek(1, io.SeekStart)
	if err != nil {
		t.Fatalf("Seek error: %v", err)
	}
	_, err = sb.Write([]byte{0xFE, 0xFF})
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	want := []byte{0x01, 0xFE, 0xFF, 0x04}
	if !bytes.Equal(sb.Bytes(), want) {
		t.Errorf("Bytes() = %v, want %v", sb.Bytes(), want)
	}
}

func TestSeekableBuffer_WriteGrowAfterSeek(t *testing.T) {
	sb := NewSeekableBuffer([]byte{0x01, 0x02})
	_, err := sb.Seek(1, io.SeekStart)
	if err != nil {
		t.Fatalf("Seek error: %v", err)
	}
	_, err = sb.Write([]byte{0xAA, 0xBB, 0xCC})
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	want := []byte{0x01, 0xAA, 0xBB, 0xCC}
	if !bytes.Equal(sb.Bytes(), want) {
		t.Errorf("Bytes() = %v, want %v", sb.Bytes(), want)
	}
}

func TestSeekableBuffer_Seek(t *testing.T) {
	tests := []struct {
		name    string
		offset  int64
		whence  int
		wantPos int64
		wantErr bool
	}{
		{"SeekStart", 2, io.SeekStart, 2, false},
		{"SeekCurrent forward", 1, io.SeekCurrent, 1, false},
		{"SeekEnd zero offset", 0, io.SeekEnd, 4, false},
		{"SeekStart zero", 0, io.SeekStart, 0, false},
		{"SeekStart negative", -1, io.SeekStart, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewSeekableBuffer([]byte{0x01, 0x02, 0x03, 0x04})
			got, err := sb.Seek(tt.offset, tt.whence)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Seek error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got != tt.wantPos {
				t.Errorf("Seek returned %d, want %d", got, tt.wantPos)
			}
		})
	}
}

func TestSeekableBuffer_RoundTrip(t *testing.T) {
	sb := NewSeekableBuffer(nil)
	want := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	_, err := sb.Write(want)
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	_, err = sb.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatalf("Seek error: %v", err)
	}
	got := make([]byte, len(want))
	_, err = io.ReadFull(sb, got)
	if err != nil {
		t.Fatalf("ReadFull error: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("round-tripped bytes = %v, want %v", got, want)
	}
}
