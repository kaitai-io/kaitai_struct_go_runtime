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
// reserved region of its parent. Complete all child writes before WriteBack;
// after success, further changes to the child are not propagated to its parent.
// Like Writer, ChildWriter is not safe for concurrent use.
type ChildWriter struct {
	*Writer

	parent    *Writer
	origin    int64
	outerSize int
	buffer    *SeekableBuffer
	encoder   ByteEncoder
	flushed   bool
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
	if innerSize < 0 || outerSize < 0 {
		return nil, fmt.Errorf("NewChild: negative size: %w", ErrBufferCapacity)
	}
	origin, err := k.Pos()
	if err != nil {
		return nil, fmt.Errorf("NewChild: %w", err)
	}
	if origin > math.MaxInt64-int64(outerSize) {
		return nil, fmt.Errorf("NewChild: reservation overflow: %w", ErrBufferCapacity)
	}
	_, err = k.Seek(origin+int64(outerSize), io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("NewChild: reserve parent region: %w", err)
	}
	buf := NewFixedSeekableBuffer(innerSize)
	child := &ChildWriter{
		Writer: NewWriter(buf), parent: k, origin: origin, outerSize: outerSize,
		buffer: buf, encoder: encoder,
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

// WriteBack flushes descendants, aligns this child's pending bits, optionally
// encodes its bytes, and writes the result at its reserved parent origin. Encoding
// failures and size mismatches leave the parent unchanged. The parent's position
// is restored after byte alignment, including on write or seek failure when its
// underlying seeker permits restoration. If both writing and restoration fail,
// the returned error wraps both causes.
//
// A failed operation may be retried; completed descendants are skipped and the
// entire parent region is overwritten on each attempt. Parent write failures may
// leave partial output; no rollback of those bytes is attempted. Success marks the
// child flushed and subsequent calls return ErrChildAlreadyWrittenBack.
func (c *ChildWriter) WriteBack() error {
	if c.flushed {
		return fmt.Errorf("WriteBack: %w", ErrChildAlreadyWrittenBack)
	}
	err := c.WriteBackChildWriters()
	if err != nil {
		return fmt.Errorf("WriteBack: descendants: %w", err)
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
	return c.writeToParent(data)
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
