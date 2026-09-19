package kaitai

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type writeOnlySeekBuffer struct {
	buffer *SeekableBuffer
}

func newWriteOnlySeekBuffer(data []byte) *writeOnlySeekBuffer {
	return &writeOnlySeekBuffer{buffer: NewSeekableBuffer(data)}
}

func (b *writeOnlySeekBuffer) Write(p []byte) (int, error) {
	return b.buffer.Write(p)
}

func (b *writeOnlySeekBuffer) Seek(offset int64, whence int) (int64, error) {
	return b.buffer.Seek(offset, whence)
}

type overlayFaultBuffer struct {
	*writeOnlySeekBuffer

	writes           int
	failWrite        int
	partialWrite     bool
	failEnd          bool
	failEndAfterMove bool
	failStart        bool
	failStartTarget  int64
}

func (b *overlayFaultBuffer) Write(p []byte) (int, error) {
	b.writes++
	if b.writes == b.failWrite {
		if b.partialWrite && len(p) > 1 {
			n, _ := b.writeOnlySeekBuffer.Write(p[:1])
			return n, io.ErrClosedPipe
		}
		return 0, io.ErrClosedPipe
	}
	return b.writeOnlySeekBuffer.Write(p)
}

func (b *overlayFaultBuffer) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekStart && b.failStart && offset == b.failStartTarget {
		return 0, io.ErrUnexpectedEOF
	}
	if whence == io.SeekEnd && b.failEnd {
		if b.failEndAfterMove {
			_, _ = b.writeOnlySeekBuffer.Seek(offset, whence)
		}
		return 0, io.ErrUnexpectedEOF
	}
	return b.writeOnlySeekBuffer.Seek(offset, whence)
}

func TestOverlayChildWritesOnlyCoveredBytes(t *testing.T) {
	buf := newWriteOnlySeekBuffer([]byte{0xaa, 0xbb, 0xcc, 0xdd})
	parent := NewWriter(buf)
	child, err := parent.NewOverlayChild(4)
	require.NoError(t, err)

	_, err = child.Seek(3, io.SeekStart)
	require.NoError(t, err)
	_, err = child.Seek(1, io.SeekStart)
	require.NoError(t, err)
	require.NoError(t, child.WriteU1(0))
	_, err = child.Seek(3, io.SeekStart)
	require.NoError(t, err)
	require.NoError(t, child.WriteU1(0xee))

	require.NoError(t, child.WriteBack())
	require.Equal(t, []byte{0xaa, 0x00, 0xcc, 0xee}, buf.buffer.Bytes())
	requireWriterPos(t, parent, 4)
}

func TestOverlayChildPreservesParentMutationsAndIgnoresParsedBytes(t *testing.T) {
	buf := NewSeekableBuffer([]byte{0, 1, 2, 3})
	parent := NewWriter(buf)
	child, err := parent.NewOverlayChild(4)
	require.NoError(t, err)
	_, err = parent.Seek(1, io.SeekStart)
	require.NoError(t, err)
	require.NoError(t, parent.WriteBytes([]byte{0xa1, 0xa2}))
	_, err = parent.Seek(4, io.SeekStart)
	require.NoError(t, err)
	_, err = child.Seek(2, io.SeekStart)
	require.NoError(t, err)
	require.NoError(t, child.WriteU1(0xf2))

	require.NoError(t, child.WriteBack())
	require.Equal(t, []byte{0, 0xa1, 0xf2, 3}, buf.Bytes())
}

func TestOverlayChildEstablishesFreshExtentWithoutClaimingHoles(t *testing.T) {
	for _, tc := range []struct {
		name  string
		write bool
		want  []byte
	}{
		{name: "empty", want: []byte{0, 0, 0, 0}},
		{name: "single-final-byte", write: true, want: []byte{0, 0, 0, 0x44}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buf := newWriteOnlySeekBuffer(nil)
			parent := NewWriter(buf)
			child, err := parent.NewOverlayChild(4)
			require.NoError(t, err)
			if tc.write {
				_, err = child.Seek(3, io.SeekStart)
				require.NoError(t, err)
				require.NoError(t, child.WriteU1(0x44))
			}
			require.NoError(t, child.WriteBack())
			require.Equal(t, tc.want, buf.buffer.Bytes())
		})
	}
}

func TestOverlayChildRequiresExtentQueryBeforePatching(t *testing.T) {
	for _, afterMove := range []bool{false, true} {
		buf := &overlayFaultBuffer{
			writeOnlySeekBuffer: newWriteOnlySeekBuffer([]byte{0xaa, 0xbb}),
			failEnd:             true,
			failEndAfterMove:    afterMove,
		}
		parent := NewWriter(buf)
		child, err := parent.NewOverlayChild(2)
		require.NoError(t, err)
		require.NoError(t, child.WriteU1(0x11))

		err = child.WriteBack()
		require.ErrorIs(t, err, io.ErrUnexpectedEOF)
		require.Equal(t, []byte{0xaa, 0xbb}, buf.buffer.Bytes())
		requireWriterPos(t, parent, 2)
	}
}

func TestOverlayChildNestedOwnership(t *testing.T) {
	buf := NewSeekableBuffer([]byte{0xaa, 0xbb, 0xcc, 0xdd})
	parent := NewWriter(buf)
	overlay, err := parent.NewOverlayChild(4)
	require.NoError(t, err)
	require.NoError(t, overlay.WriteU1(0x11))
	nestedOverlay, err := overlay.NewOverlayChild(2)
	require.NoError(t, err)
	_, err = nestedOverlay.Seek(1, io.SeekStart)
	require.NoError(t, err)
	require.NoError(t, nestedOverlay.WriteU1(0x22))
	dense, err := overlay.NewChild(1, 1, nil)
	require.NoError(t, err)

	require.NoError(t, overlay.WriteBack())
	require.Equal(t, []byte{0x11, 0xbb, 0x22, 0x00}, buf.Bytes())
	require.True(t, nestedOverlay.flushed)
	require.True(t, dense.flushed)
}

func TestOverlayChildrenUseDeterministicLaterSiblingPrecedence(t *testing.T) {
	buf := NewSeekableBuffer([]byte{0xaa, 0xbb, 0xcc, 0xdd})
	parent := NewWriter(buf)
	first, err := parent.NewOverlayChild(4)
	require.NoError(t, err)
	_, err = first.Seek(1, io.SeekStart)
	require.NoError(t, err)
	require.NoError(t, first.WriteBytes([]byte{0x11, 0x12}))
	_, err = parent.Seek(1, io.SeekStart)
	require.NoError(t, err)
	second, err := parent.NewOverlayChild(2)
	require.NoError(t, err)
	require.NoError(t, second.WriteU1(0xee))
	_, err = parent.Seek(4, io.SeekStart)
	require.NoError(t, err)

	require.NoError(t, parent.WriteBackChildWriters())
	require.Equal(t, []byte{0xaa, 0xee, 0x12, 0xdd}, buf.Bytes())
}

func TestOverlayChildOwnsAlignedPartialBitByte(t *testing.T) {
	for _, tc := range []struct {
		name  string
		write func(*Writer) error
		want  byte
	}{
		{name: "big-endian", write: func(w *Writer) error { return w.WriteBitsIntBe(3, 5) }, want: 0xa0},
		{name: "little-endian", write: func(w *Writer) error { return w.WriteBitsIntLe(3, 5) }, want: 0x05},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buf := NewSeekableBuffer([]byte{0xff})
			child, err := NewWriter(buf).NewOverlayChild(1)
			require.NoError(t, err)
			require.NoError(t, tc.write(child.Writer))
			require.NoError(t, child.WriteBack())
			require.Equal(t, []byte{tc.want}, buf.Bytes())
		})
	}
}

func TestOverlayChildRetryReplaysPartiallyWrittenRanges(t *testing.T) {
	buf := &overlayFaultBuffer{
		writeOnlySeekBuffer: newWriteOnlySeekBuffer([]byte{0xaa, 0xbb, 0xcc, 0xdd}),
		failWrite:           1,
		partialWrite:        true,
	}
	parent := NewWriter(buf)
	child, err := parent.NewOverlayChild(4)
	require.NoError(t, err)
	require.NoError(t, child.WriteBytes([]byte{0x10, 0x11}))
	_, err = child.Seek(3, io.SeekStart)
	require.NoError(t, err)
	require.NoError(t, child.WriteU1(0x13))

	err = child.WriteBack()
	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Equal(t, []byte{0x10, 0xbb, 0xcc, 0xdd}, buf.buffer.Bytes())
	requireWriterPos(t, parent, 4)
	buf.failWrite = 0
	require.NoError(t, child.WriteBack())
	require.Equal(t, []byte{0x10, 0x11, 0xcc, 0x13}, buf.buffer.Bytes())
}

func TestOverlayChildRetryReplaysEarlierRangesAfterMiddleFailure(t *testing.T) {
	buf := &overlayFaultBuffer{
		writeOnlySeekBuffer: newWriteOnlySeekBuffer([]byte{0xaa, 0xbb, 0xcc, 0xdd}),
		failWrite:           2,
	}
	parent := NewWriter(buf)
	child, err := parent.NewOverlayChild(4)
	require.NoError(t, err)
	require.NoError(t, child.WriteBytes([]byte{0x10, 0x11}))
	_, err = child.Seek(3, io.SeekStart)
	require.NoError(t, err)
	require.NoError(t, child.WriteU1(0x13))

	require.ErrorIs(t, child.WriteBack(), io.ErrClosedPipe)
	require.Equal(t, []byte{0x10, 0x11, 0xcc, 0xdd}, buf.buffer.Bytes())
	requireWriterPos(t, parent, 4)
	buf.failWrite = 0
	require.NoError(t, child.WriteBack())
	require.Equal(t, []byte{0x10, 0x11, 0xcc, 0x13}, buf.buffer.Bytes())
}

func TestOverlayChildRestoreFailureIsJoinedAndRetryable(t *testing.T) {
	buf := &overlayFaultBuffer{writeOnlySeekBuffer: newWriteOnlySeekBuffer([]byte{0xaa, 0xbb})}
	parent := NewWriter(buf)
	child, err := parent.NewOverlayChild(2)
	require.NoError(t, err)
	require.NoError(t, child.WriteBytes([]byte{0x10, 0x11}))
	buf.failStart, buf.failStartTarget = true, 2

	require.ErrorIs(t, child.WriteBack(), io.ErrUnexpectedEOF)
	require.Equal(t, []byte{0x10, 0x11}, buf.buffer.Bytes())
	require.False(t, child.flushed)
	buf.failStart = false
	require.NoError(t, child.WriteBack())
	requireWriterPos(t, parent, 2)
}

func TestOverlayChildRetriesPartialExtentCreation(t *testing.T) {
	buf := &overlayFaultBuffer{
		writeOnlySeekBuffer: newWriteOnlySeekBuffer(nil),
		failWrite:           1,
		partialWrite:        true,
	}
	parent := NewWriter(buf)
	child, err := parent.NewOverlayChild(3)
	require.NoError(t, err)
	_, err = child.Seek(2, io.SeekStart)
	require.NoError(t, err)
	require.NoError(t, child.WriteU1(0x33))

	require.ErrorIs(t, child.WriteBack(), io.ErrClosedPipe)
	require.Equal(t, []byte{0}, buf.buffer.Bytes())
	requireWriterPos(t, parent, 3)
	buf.failWrite = 0
	require.NoError(t, child.WriteBack())
	require.Equal(t, []byte{0, 0, 0x33}, buf.buffer.Bytes())
}

func TestOverlayChildRetryDoesNotRepeatSuccessfulDenseDescendant(t *testing.T) {
	buf := &overlayFaultBuffer{
		writeOnlySeekBuffer: newWriteOnlySeekBuffer([]byte{0xaa, 0xbb}),
		failWrite:           1,
	}
	parent := NewWriter(buf)
	overlay, err := parent.NewOverlayChild(2)
	require.NoError(t, err)
	encoderCalls := 0
	dense, err := overlay.NewChild(2, 2, testByteEncoder(func(data []byte) ([]byte, error) {
		encoderCalls++
		return bytes.Clone(data), nil
	}))
	require.NoError(t, err)
	require.NoError(t, dense.WriteBytes([]byte{0x11, 0x22}))

	require.ErrorIs(t, overlay.WriteBack(), io.ErrClosedPipe)
	buf.failWrite = 0
	require.NoError(t, overlay.WriteBack())
	require.Equal(t, 1, encoderCalls)
	require.Equal(t, []byte{0x11, 0x22}, buf.buffer.Bytes())
}

func TestOverlayChildPatchesDenseProcessedParentBeforeEncoding(t *testing.T) {
	buf := NewSeekableBuffer(nil)
	root := NewWriter(buf)
	dense, err := root.NewChild(3, 3, testByteEncoder(func(data []byte) ([]byte, error) {
		for i := range data {
			data[i] ^= 0xff
		}
		return data, nil
	}))
	require.NoError(t, err)
	require.NoError(t, dense.WriteBytes([]byte{1, 2, 3}))
	_, err = dense.Seek(1, io.SeekStart)
	require.NoError(t, err)
	overlay, err := dense.NewOverlayChild(1)
	require.NoError(t, err)
	require.NoError(t, overlay.WriteU1(9))

	require.NoError(t, dense.WriteBack())
	require.Equal(t, []byte{0xfe, 0xf6, 0xfc}, buf.Bytes())
}

func TestOverlayChildWriteBackNoAlignPreservesParentBits(t *testing.T) {
	buf := NewSeekableBuffer(nil)
	parent := NewWriter(buf)
	child, err := parent.NewOverlayChild(1)
	require.NoError(t, err)
	require.NoError(t, child.WriteU1(0x42))
	require.NoError(t, parent.WriteBitsIntBe(3, 5))
	physical, err := buf.Seek(0, io.SeekCurrent)
	require.NoError(t, err)

	require.NoError(t, child.WriteBackNoAlign())
	got, err := buf.Seek(0, io.SeekCurrent)
	require.NoError(t, err)
	require.Equal(t, physical, got)
	require.Equal(t, 3, parent.bitsLeft)
	require.Equal(t, uint64(5), parent.bits)
	require.NoError(t, parent.WriteBitsIntBe(5, 19))
	require.Equal(t, []byte{0x42, 0xb3}, buf.Bytes())
}

func TestOverlayChildFixedReservationFailureIsAtomic(t *testing.T) {
	buf := NewFixedSeekableBuffer(2)
	parent := NewWriter(buf)
	_, err := parent.Seek(1, io.SeekStart)
	require.NoError(t, err)
	_, err = parent.NewOverlayChild(2)
	require.ErrorIs(t, err, ErrBufferCapacity)
	requireWriterPos(t, parent, 1)
	require.Equal(t, []byte{0, 0}, buf.Bytes())
}

func TestOverlayChildExtentFailureCanBeRetried(t *testing.T) {
	buf := &overlayFaultBuffer{
		writeOnlySeekBuffer: newWriteOnlySeekBuffer(nil),
		failEnd:             true,
	}
	parent := NewWriter(buf)
	child, err := parent.NewOverlayChild(2)
	require.NoError(t, err)
	_, err = child.Seek(1, io.SeekStart)
	require.NoError(t, err)
	require.NoError(t, child.WriteU1(0x22))
	require.ErrorIs(t, child.WriteBack(), io.ErrUnexpectedEOF)
	require.Empty(t, buf.buffer.Bytes())
	buf.failEnd = false
	require.NoError(t, child.WriteBack())
	require.Equal(t, []byte{0, 0x22}, buf.buffer.Bytes())
}
