package kaitai

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type testByteEncoder func([]byte) ([]byte, error)

func (f testByteEncoder) Encode(data []byte) ([]byte, error) { return f(data) }

func requireWriterPos(t *testing.T, w *Writer, want int64) {
	t.Helper()
	got, err := w.Pos()
	if err != nil || got != want {
		t.Fatalf("Pos() = (%d, %v), want (%d, nil)", got, err, want)
	}
}

func TestChildWriterReservationAndWriteBack(t *testing.T) {
	buf := NewSeekableBuffer(nil)
	parent := NewWriter(buf)
	require.NoError(t, parent.WriteU1(0xaa))
	child, err := parent.NewChild(3, 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	requireWriterPos(t, parent, 4)
	requireWriterPos(t, child.Writer, 0)
	require.NoError(t, child.WriteBytes([]byte{0x11, 0x22}))
	require.NoError(t, parent.WriteU1(0xbb))
	require.NoError(t, child.WriteBack())
	if !bytes.Equal(buf.Bytes(), []byte{0xaa, 0x11, 0x22, 0, 0xbb}) {
		t.Fatalf("bytes = %x, want aa112200bb", buf.Bytes())
	}
	requireWriterPos(t, parent, 5)
	requireWriterPos(t, child.Writer, 2)
	require.ErrorIs(t, child.WriteBack(), ErrChildAlreadyWrittenBack)
}

func TestChildWriterOverflowDoesNotPartiallyWrite(t *testing.T) {
	buf := NewSeekableBuffer(nil)
	parent := NewWriter(buf)
	child, err := parent.NewChild(2, 2, nil)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, child.WriteU1(0x11))
	require.ErrorIs(t, child.WriteBytes([]byte{0x22, 0x33}), ErrBufferCapacity)
	requireWriterPos(t, child.Writer, 1)
	require.NoError(t, child.WriteBack())
	if !bytes.Equal(buf.Bytes(), []byte{0x11, 0}) {
		t.Fatalf("bytes = %x", buf.Bytes())
	}
}

func TestChildWriterFailedReservation(t *testing.T) {
	parent := NewWriter(NewFixedSeekableBuffer(2))
	for _, size := range []int{-1, 3} {
		_, err := parent.NewChild(size, size, nil)
		if !errors.Is(err, ErrBufferCapacity) {
			t.Fatalf("NewChild(%d) = %v", size, err)
		}
		requireWriterPos(t, parent, 0)
	}
	require.NoError(t, parent.WriteBackChildWriters())
}

func TestChildWriterNonSeekableRootStillWrites(t *testing.T) {
	var buf bytes.Buffer
	parent := NewWriter(&buf)
	require.NoError(t, parent.WriteBitsIntBe(3, 5))
	_, err := parent.NewChild(2, 2, nil)
	if !errors.Is(err, ErrWriterNotSeekable) {
		t.Fatalf("NewChild() = %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("failed reservation flushed bits: %x", buf.Bytes())
	}
	require.NoError(t, parent.WriteU1(0x11))
	require.NoError(t, parent.WriteBackChildWriters())
	if !bytes.Equal(buf.Bytes(), []byte{0xa0, 0x11}) {
		t.Fatalf("bytes = %x", buf.Bytes())
	}
}

func TestWriteBackNestedDepthFirstAndPositions(t *testing.T) {
	buf := NewSeekableBuffer(nil)
	parent := NewWriter(buf)
	encoder := testByteEncoder(func(data []byte) ([]byte, error) {
		if !bytes.Equal(data, []byte{0x11, 0x22, 0x33, 0x44}) {
			t.Fatalf("encoder saw unfinished children: %x", data)
		}
		return data, nil
	})
	child, err := parent.NewChild(4, 4, encoder)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, child.WriteU1(0x11))
	grandchild, err := child.NewChild(2, 2, nil)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, grandchild.WriteBytes([]byte{0x22, 0x33}))
	require.NoError(t, child.WriteU1(0x44))
	_, err = child.Seek(1, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, parent.WriteU1(0x55))
	require.NoError(t, parent.WriteBackChildWriters())
	if !bytes.Equal(buf.Bytes(), []byte{0x11, 0x22, 0x33, 0x44, 0x55}) {
		t.Fatalf("bytes = %x", buf.Bytes())
	}
	requireWriterPos(t, parent, 5)
	requireWriterPos(t, child.Writer, 1)
	requireWriterPos(t, grandchild.Writer, 2)
	require.NoError(t, parent.WriteBackChildWriters())
}

func TestWriteBackEncoderFailureAndRetry(t *testing.T) {
	buf := NewSeekableBuffer([]byte{0xaa, 0xbb})
	parent := NewWriter(buf)
	wantErr := io.ErrUnexpectedEOF
	fail := true
	child, err := parent.NewChild(2, 2, testByteEncoder(func(data []byte) ([]byte, error) {
		data[0] ^= 0xff
		if fail {
			return nil, wantErr
		}
		return data, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, child.WriteBytes([]byte{0x12, 0x34}))
	require.NoError(t, parent.WriteBitsIntBe(3, 5))
	require.ErrorIs(t, child.WriteBack(), wantErr)
	if !bytes.Equal(buf.Bytes(), []byte{0xaa, 0xbb}) {
		t.Fatalf("parent mutated: %x", buf.Bytes())
	}
	requireWriterPos(t, parent, 3)
	fail = false
	require.NoError(t, child.WriteBack())
	if !bytes.Equal(buf.Bytes(), []byte{0xed, 0x34, 0xa0}) {
		t.Fatalf("encoder input or pending parent bits corrupted: %x", buf.Bytes())
	}
	requireWriterPos(t, parent, 3)
}

func TestWriteBackEncoderSize(t *testing.T) {
	for _, tc := range []struct {
		name    string
		encoded []byte
	}{
		{"short", []byte{0x33}},
		{"long", []byte{1, 2, 3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buf := NewSeekableBuffer([]byte{0xaa, 0xaa, 0xbb})
			parent := NewWriter(buf)
			child, err := parent.NewChild(2, 2, testByteEncoder(func([]byte) ([]byte, error) { return tc.encoded, nil }))
			if err != nil {
				t.Fatal(err)
			}
			err = child.WriteBack()
			var sizeErr ConsistencyError
			if !errors.As(err, &sizeErr) || sizeErr.Expected() != 2 || sizeErr.Actual() != len(tc.encoded) {
				t.Fatalf("error = %v", err)
			}
			if !bytes.Equal(buf.Bytes(), []byte{0xaa, 0xaa, 0xbb}) {
				t.Fatalf("bytes = %x", buf.Bytes())
			}
			requireWriterPos(t, parent, 2)
		})
	}
}

func TestWriteBackAlignsChildAndParentBits(t *testing.T) {
	buf := NewSeekableBuffer(nil)
	parent := NewWriter(buf)
	require.NoError(t, parent.WriteBitsIntBe(3, 5))
	child, err := parent.NewChild(1, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	requireWriterPos(t, parent, 2)
	require.NoError(t, child.WriteBitsIntBe(3, 3))
	require.NoError(t, parent.WriteBitsIntLe(3, 5))
	require.NoError(t, child.WriteBack())
	if !bytes.Equal(buf.Bytes(), []byte{0xa0, 0x60, 0x05}) {
		t.Fatalf("bytes = %x", buf.Bytes())
	}
	requireWriterPos(t, parent, 3)
	requireWriterPos(t, child.Writer, 1)
}

type writeBackFaultBuffer struct {
	*SeekableBuffer

	writeErr   error
	seekErr    error
	seekTarget int64
}

func (b *writeBackFaultBuffer) Write(data []byte) (int, error) {
	if b.writeErr != nil {
		return 0, b.writeErr
	}
	return b.SeekableBuffer.Write(data)
}

func (b *writeBackFaultBuffer) Seek(offset int64, whence int) (int64, error) {
	if b.seekErr != nil && whence == io.SeekStart && offset == b.seekTarget {
		return 0, b.seekErr
	}
	return b.SeekableBuffer.Seek(offset, whence)
}

func TestWriteBackParentFailureAndRetry(t *testing.T) {
	for _, tc := range []struct {
		name              string
		writeErr, seekErr error
		seekTarget        int64
		checkPos          bool
	}{
		{"write", io.ErrClosedPipe, nil, 0, true},
		{"seek", nil, io.ErrUnexpectedEOF, 0, true},
		{"restore", nil, io.ErrUnexpectedEOF, 2, false},
		{"write-and-restore", io.ErrClosedPipe, io.ErrUnexpectedEOF, 2, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buf := &writeBackFaultBuffer{SeekableBuffer: NewSeekableBuffer([]byte{0xaa, 0xbb})}
			parent := NewWriter(buf)
			child, err := parent.NewChild(2, 2, nil)
			require.NoError(t, err)
			require.NoError(t, child.WriteBytes([]byte{0x12, 0x34}))
			buf.writeErr, buf.seekErr, buf.seekTarget = tc.writeErr, tc.seekErr, tc.seekTarget
			err = child.WriteBack()
			if buf.writeErr != nil && !errors.Is(err, tc.writeErr) {
				t.Fatalf("missing write error: %v", err)
			}
			if buf.seekErr != nil && !errors.Is(err, tc.seekErr) {
				t.Fatalf("missing seek error: %v", err)
			}
			if tc.checkPos {
				requireWriterPos(t, parent, 2)
			}
			buf.writeErr, buf.seekErr = nil, nil
			require.NoError(t, child.WriteBack())
			if !bytes.Equal(buf.Bytes(), []byte{0x12, 0x34}) {
				t.Fatalf("retry bytes = %x", buf.Bytes())
			}
		})
	}
}

func TestWriteBackRetryDoesNotRepeatSuccessfulNestedChildren(t *testing.T) {
	buf := &writeBackFaultBuffer{SeekableBuffer: NewSeekableBuffer(nil)}
	parent := NewWriter(buf)
	child, err := parent.NewChild(2, 2, nil)
	if err != nil {
		t.Fatal(err)
	}
	grandchild, err := child.NewChild(2, 2, nil)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, grandchild.WriteBytes([]byte{0x12, 0x34}))
	buf.writeErr = io.ErrClosedPipe
	require.ErrorIs(t, parent.WriteBackChildWriters(), buf.writeErr)
	buf.writeErr = nil
	require.NoError(t, parent.WriteBackChildWriters())
	if !bytes.Equal(buf.Bytes(), []byte{0x12, 0x34}) {
		t.Fatalf("bytes = %x", buf.Bytes())
	}
}

func TestWriteBackAlignmentFailurePreservesPendingBits(t *testing.T) {
	buf := &writeBackFaultBuffer{SeekableBuffer: NewSeekableBuffer(nil)}
	parent := NewWriter(buf)
	child, err := parent.NewChild(1, 1, nil)
	require.NoError(t, err)
	require.NoError(t, child.WriteU1(0x12))
	require.NoError(t, parent.WriteBitsIntBe(3, 5))
	buf.writeErr = io.ErrClosedPipe
	require.ErrorIs(t, child.WriteBack(), io.ErrClosedPipe)
	requireWriterPos(t, parent, 2)
	buf.writeErr = nil
	require.NoError(t, child.WriteBack())
	require.Equal(t, []byte{0x12, 0xa0}, buf.Bytes())
	requireWriterPos(t, parent, 2)
}

func TestChildWriterBitCapacityFailureIsImmediate(t *testing.T) {
	buf := NewSeekableBuffer(nil)
	child, err := NewWriter(buf).NewChild(1, 1, nil)
	require.NoError(t, err)
	require.NoError(t, child.WriteU1(0x12))
	require.ErrorIs(t, child.WriteBitsIntBe(3, 5), ErrBufferCapacity)
	require.Empty(t, buf.Bytes())
}

func TestChildWriterDistinctInnerAndOuterSizes(t *testing.T) {
	for _, tc := range []struct {
		name           string
		input, encoded []byte
	}{
		{"expand", []byte{0x12}, []byte{0x12, 0x34, 0x56}},
		{"shrink", []byte{0x12, 0x34, 0x56}, []byte{0x78}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buf := NewSeekableBuffer(nil)
			parent := NewWriter(buf)
			child, err := parent.NewChild(len(tc.input), len(tc.encoded), testByteEncoder(func(data []byte) ([]byte, error) {
				if !bytes.Equal(data, tc.input) {
					t.Fatalf("logical bytes = %x, want %x", data, tc.input)
				}
				return tc.encoded, nil
			}))
			if err != nil {
				t.Fatal(err)
			}
			requireWriterPos(t, parent, int64(len(tc.encoded)))
			require.NoError(t, child.WriteBytes(tc.input))
			require.ErrorIs(t, child.WriteU1(0xff), ErrBufferCapacity)
			require.NoError(t, child.WriteBack())
			if !bytes.Equal(buf.Bytes(), tc.encoded) {
				t.Fatalf("encoded bytes = %x, want %x", buf.Bytes(), tc.encoded)
			}
		})
	}
}
