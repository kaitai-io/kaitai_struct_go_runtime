package kaitai

import (
	"bytes"
	"errors"
	"io"
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
