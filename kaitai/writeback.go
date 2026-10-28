package kaitai

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
)

// ByteEncoder transforms logical bytes into their on-disk representation.
// Encode may modify its input, which is a copy of the child buffer. A failed
// write-back may call Encode again; implementations should be deterministic.
type ByteEncoder interface {
	Encode(data []byte) ([]byte, error)
}

// ErrChildAlreadyWrittenBack reports an attempt to write back a flushed child.
var ErrChildAlreadyWrittenBack = errors.New("child writer already written back")

// ChildWriter buffers a fixed-size logical stream for explicit write-back to a
// reserved region of its parent. A child created by NewChild owns its complete
// reservation; a child created by NewOverlayChild owns only bytes explicitly
// written to its logical stream. Complete all child writes before WriteBack;
// after success, further changes to the child are not propagated to its parent.
// Like Writer, ChildWriter is not safe for concurrent use.
type ChildWriter struct {
	*Writer

	parent    *Writer
	origin    int64
	outerSize int
	buffer    *SeekableBuffer
	encoder   ByteEncoder
	coverage  []uint64
	overlay   bool
	flushed   bool
}

type overlayBuffer struct {
	*SeekableBuffer

	coverage []uint64
}

func (b *overlayBuffer) Write(data []byte) (int, error) {
	start := b.pos
	n, err := b.SeekableBuffer.Write(data)
	for offset := start; offset < start+int64(n); offset++ {
		word, bit := offset/64, offset%64
		b.coverage[word] |= uint64(1) << bit
	}
	return n, err
}

// NewChild reserves outerSize bytes at the current position and creates a
// zero-filled child with innerSize bytes of fixed capacity. The parent's pending
// bits are aligned before reservation; the child starts independently at zero.
// Reservation advances the parent without writing the reserved bytes.
// Negative sizes and reservations outside a fixed parent return ErrBufferCapacity.
// A seekable parent is required. A nil encoder leaves logical bytes unchanged.
// Encoded output must exactly match outerSize; schema-specific padding and
// terminators, if any, must be supplied by the encoder.
func (k *Writer) NewChild(innerSize, outerSize int, encoder ByteEncoder) (*ChildWriter, error) {
	return k.newChild(innerSize, outerSize, encoder, false)
}

// NewOverlayChild reserves size bytes at the current position and creates a
// sparse fixed-size child. Only bytes explicitly written by the child are
// propagated to existing parent bytes; seeking alone does not claim bytes.
// If the reservation extends past the parent's physical end, finalization
// zero-fills only the newly created extent. The parent need only implement
// io.WriteSeeker; it is never read.
func (k *Writer) NewOverlayChild(size int) (*ChildWriter, error) {
	return k.newChild(size, size, nil, true)
}

func (k *Writer) newChild(innerSize, outerSize int, encoder ByteEncoder, overlay bool) (*ChildWriter, error) {
	operation := "NewChild"
	if overlay {
		operation = "NewOverlayChild"
	}
	if innerSize < 0 || outerSize < 0 {
		return nil, fmt.Errorf("%s: negative size: %w", operation, ErrBufferCapacity)
	}
	origin, err := k.Pos()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	if origin > math.MaxInt64-int64(outerSize) {
		return nil, fmt.Errorf("%s: reservation overflow: %w", operation, ErrBufferCapacity)
	}
	_, err = k.Seek(origin+int64(outerSize), io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("%s: reserve parent region: %w", operation, err)
	}
	buf := NewFixedSeekableBuffer(innerSize)
	writer := NewWriter(buf)
	var coverage []uint64
	if overlay {
		words := innerSize / 64
		if innerSize%64 != 0 {
			words++
		}
		coverage = make([]uint64, words)
		writer = NewWriter(&overlayBuffer{SeekableBuffer: buf, coverage: coverage})
	}
	child := &ChildWriter{
		Writer: writer, parent: k, origin: origin, outerSize: outerSize,
		buffer: buf, encoder: encoder, coverage: coverage, overlay: overlay,
	}
	k.children = append(k.children, child)
	return child, nil
}

// WriteBackChildWriters flushes registered children depth-first, preserving this
// writer's position after byte alignment. Successfully flushed children are
// skipped, so a failed traversal can be retried and an empty traversal is a no-op.
func (k *Writer) WriteBackChildWriters() error {
	for _, child := range k.children {
		if child.flushed {
			continue
		}
		err := child.WriteBack()
		if err != nil {
			return err
		}
	}
	k.children = nil
	return nil
}

// WriteBack flushes descendants and aligns this child's pending bits. A dense
// child created by NewChild optionally encodes its logical bytes and overwrites
// its complete parent reservation. An overlay child created by NewOverlayChild
// first establishes any reservation extent missing beyond the parent's physical
// end, then writes only its covered byte ranges and preserves existing bytes in
// uncovered ranges. Encoding failures and size mismatches leave the parent
// unchanged. The parent's position is restored after byte alignment, including
// on write or seek failure when its underlying seeker permits restoration. If
// both writing and restoration fail, the returned error wraps both causes.
//
// A failed operation may be retried; completed descendants are skipped and the
// dense child rewrites its complete reservation. An overlay retry re-establishes
// any still-missing extent and replays all covered ranges; it never claims an
// existing uncovered byte. Parent write failures may leave partial output; no
// rollback of those bytes is attempted. Success marks the child flushed and
// subsequent calls return ErrChildAlreadyWrittenBack.
func (c *ChildWriter) WriteBack() error {
	return c.writeBack(false)
}

// WriteBackNoAlign propagates a reserved child while preserving the parent's
// pending partial byte and physical cursor. This is used by generated output
// sessions whose root may end with bits for the caller to continue. Descendants
// are completed depth-first. Direct WriteBack retains its alignment contract.
func (c *ChildWriter) WriteBackNoAlign() error {
	return c.writeBack(true)
}

func (c *ChildWriter) writeBack(noAlign bool) error {
	if c.flushed {
		return fmt.Errorf("WriteBack: %w", ErrChildAlreadyWrittenBack)
	}
	err := c.writeBackChildren(noAlign)
	if err != nil {
		return err
	}
	err = c.AlignToByte()
	if err != nil {
		return fmt.Errorf("WriteBack: align child: %w", err)
	}
	data := bytes.Clone(c.buffer.Bytes())
	if c.encoder != nil {
		data, err = c.encoder.Encode(data)
		if err != nil {
			return fmt.Errorf("WriteBack: encode: %w", err)
		}
	}
	if len(data) != c.outerSize {
		return fmt.Errorf("WriteBack: %w", NewConsistencyError("write-back size", c.outerSize, len(data)))
	}
	if noAlign {
		bits, left, le, reserved := c.parent.bits, c.parent.bitsLeft, c.parent.bitsLe, c.parent.bitsReserved
		c.parent.bits, c.parent.bitsLeft, c.parent.bitsLe, c.parent.bitsReserved = 0, 0, false, false
		defer func() {
			c.parent.bits, c.parent.bitsLeft, c.parent.bitsLe, c.parent.bitsReserved = bits, left, le, reserved
		}()
	}
	if c.overlay {
		return c.writeOverlayToParent(data)
	}
	return c.writeToParent(data)
}

func (c *ChildWriter) writeBackChildren(noAlign bool) error {
	for _, child := range c.children {
		if !child.flushed {
			err := child.writeBack(noAlign)
			if err != nil {
				return fmt.Errorf("WriteBack: descendants: %w", err)
			}
		}
	}
	c.children = nil
	return nil
}

func (c *ChildWriter) writeToParent(data []byte) (err error) {
	err = c.parent.AlignToByte()
	if err != nil {
		return fmt.Errorf("WriteBack: align parent: %w", err)
	}
	pos, err := c.parent.Pos()
	if err != nil {
		return fmt.Errorf("WriteBack: save parent position: %w", err)
	}
	defer func() {
		_, restoreErr := c.parent.Seek(pos, io.SeekStart)
		if restoreErr != nil {
			err = errors.Join(err, fmt.Errorf("WriteBack: restore parent position: %w", restoreErr))
		}
		if err == nil {
			c.flushed = true
		}
	}()
	_, err = c.parent.Seek(c.origin, io.SeekStart)
	if err != nil {
		return fmt.Errorf("WriteBack: seek parent origin: %w", err)
	}
	err = c.parent.WriteBytes(data)
	if err != nil {
		return fmt.Errorf("WriteBack: write parent: %w", err)
	}
	return nil
}

func (c *ChildWriter) writeOverlayToParent(data []byte) (err error) {
	err = c.parent.AlignToByte()
	if err != nil {
		return fmt.Errorf("WriteBack: align parent: %w", err)
	}
	pos, err := c.parent.Pos()
	if err != nil {
		return fmt.Errorf("WriteBack: save parent position: %w", err)
	}
	defer func() {
		_, restoreErr := c.parent.Seek(pos, io.SeekStart)
		if restoreErr != nil {
			err = errors.Join(err, fmt.Errorf("WriteBack: restore parent position: %w", restoreErr))
		}
		if err == nil {
			c.flushed = true
		}
	}()

	err = c.ensureOverlayExtent()
	if err != nil {
		return err
	}
	err = c.writeOverlayRanges(data)
	return err
}

func (c *ChildWriter) ensureOverlayExtent() error {
	physicalEnd, err := c.parent.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("WriteBack: query parent extent: %w", err)
	}
	reservationEnd := c.origin + int64(c.outerSize)
	if physicalEnd >= reservationEnd {
		return nil
	}
	start := max(physicalEnd, c.origin)
	_, err = c.parent.Seek(start, io.SeekStart)
	if err != nil {
		return fmt.Errorf("WriteBack: seek parent extension: %w", err)
	}
	var zeroes [4096]byte
	for remaining := reservationEnd - start; remaining > 0; {
		n := min(remaining, int64(len(zeroes)))
		err = c.parent.WriteBytes(zeroes[:n])
		if err != nil {
			return fmt.Errorf("WriteBack: extend parent: %w", err)
		}
		remaining -= n
	}
	return nil
}

func (c *ChildWriter) writeOverlayRanges(data []byte) error {
	for start := 0; start < c.outerSize; {
		for start < c.outerSize && !c.covers(start) {
			start++
		}
		if start == c.outerSize {
			break
		}
		end := start + 1
		for end < c.outerSize && c.covers(end) {
			end++
		}
		_, err := c.parent.Seek(c.origin+int64(start), io.SeekStart)
		if err != nil {
			return fmt.Errorf("WriteBack: seek parent overlay: %w", err)
		}
		err = c.parent.WriteBytes(data[start:end])
		if err != nil {
			return fmt.Errorf("WriteBack: write parent overlay: %w", err)
		}
		start = end
	}
	return nil
}

func (c *ChildWriter) covers(offset int) bool {
	word, bit := offset/64, offset%64
	return c.coverage[word]&(uint64(1)<<bit) != 0
}
