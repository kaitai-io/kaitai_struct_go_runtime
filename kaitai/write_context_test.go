package kaitai

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteContextCompletesNestedOperationsBeforeNextSibling(t *testing.T) {
	c := NewWriteContext()
	defer c.Close()
	var order []int
	root, nested, late := new(int), new(int), new(int)
	leave, err := c.Enter(root, func() { order = append(order, 6) })
	if err != nil {
		t.Fatal(err)
	}
	leaveChild, err := c.Enter(nested, func() { order = append(order, 5) })
	if err != nil {
		t.Fatal(err)
	}
	c.Defer(func() error { order = append(order, 1); return nil })
	leaveChild()
	c.Defer(func() error {
		order = append(order, 2)
		leaveLate, err := c.Enter(late, func() { order = append(order, 4) })
		if err != nil {
			return err
		}
		defer leaveLate()
		c.Defer(func() error { order = append(order, 3); return nil })
		return nil
	})
	leave()
	require.NoError(t, c.Finish())
	c.Close()
	if !reflect.DeepEqual(order, []int{1, 2, 3, 4, 5, 6}) {
		t.Fatalf("order %v", order)
	}
	_, err = c.Enter(new(int), func() {})
	require.ErrorIs(t, err, ErrWriteContextClosed)
}

func TestWriteBackNoAlignFailureAndRetryPreservesPhysicalCursorAndBits(t *testing.T) {
	for _, overlap := range []bool{false, true} {
		buf := &writeBackFaultBuffer{SeekableBuffer: NewSeekableBuffer(nil)}
		parent := NewWriter(buf)
		calls := 0
		child, err := parent.NewChild(1, 1, ByteEncoderFunc(func(data []byte) ([]byte, error) {
			calls++
			return data, nil
		}))
		require.NoError(t, err)
		require.NoError(t, child.WriteU1(0x42))
		if overlap {
			_, err = parent.Seek(0, io.SeekStart)
			require.NoError(t, err)
		}
		require.NoError(t, parent.WriteBitsIntBe(3, 5))
		physical, err := buf.Seek(0, io.SeekCurrent)
		require.NoError(t, err)
		buf.writeErr = io.ErrClosedPipe
		require.ErrorIs(t, child.WriteBackNoAlign(), io.ErrClosedPipe)
		got, err := buf.Seek(0, io.SeekCurrent)
		require.NoError(t, err)
		require.Equal(t, physical, got)
		if parent.bitsLeft != 3 || parent.bits != 5 {
			t.Fatalf("bits lost %d/%d", parent.bits, parent.bitsLeft)
		}
		buf.writeErr = nil
		require.NoError(t, child.WriteBackNoAlign())
		if calls != 2 {
			t.Fatalf("encoder retry calls %d", calls)
		}
		require.NoError(t, parent.WriteBitsIntBe(5, 19))
		want := []byte{0x42, 0xb3}
		if overlap {
			want = []byte{0xb3}
		}
		if !bytes.Equal(buf.Bytes(), want) {
			t.Fatalf("output %x", buf.Bytes())
		}
	}
}

func TestWriteContextOwnsOnlyCurrentReservationsAndPreservesBits(t *testing.T) {
	buf := NewSeekableBuffer(nil)
	w := NewWriter(buf)
	old, err := w.NewChild(1, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, old.WriteU1(0xee))
	c := NewWriteContext()
	defer c.Close()
	child, err := w.NewChild(1, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	c.RegisterChild(child)
	require.NoError(t, child.WriteU1(0x42))
	require.NoError(t, w.WriteBitsIntBe(3, 5))
	require.NoError(t, c.Finish())
	if got := buf.Bytes(); !bytes.Equal(got, []byte{0, 0x42}) {
		t.Fatalf("premature alignment or historical flush: %x", got)
	}
	require.NoError(t, w.WriteBitsIntBe(5, 0x13))
	if got := buf.Bytes(); !bytes.Equal(got, []byte{0, 0x42, 0xb3}) {
		t.Fatalf("bits lost: %x", got)
	}
	if old.flushed {
		t.Fatal("historical child flushed")
	}
}

func TestWriteContextFailureRestoresWithoutFlushingAndRejectsDuplicateBinding(t *testing.T) {
	c := NewWriteContext()
	object := new(int)
	restored := false
	leave, err := c.Enter(object, func() { restored = true })
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Enter(object, func() {})
	require.Error(t, err)
	w := NewWriter(NewSeekableBuffer(nil))
	child, err := w.NewChild(1, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	c.RegisterChild(child)
	want := io.ErrClosedPipe
	c.Defer(func() error { return want })
	leave()
	require.ErrorIs(t, c.Finish(), want)
	c.Close()
	if !restored || child.flushed || len(w.children) != 0 {
		t.Fatalf("cleanup restored=%v flushed=%v children=%d", restored, child.flushed, len(w.children))
	}
	require.NoError(t, child.WriteBack())
}

func TestWriteContextRunBoundsReentryPermissionAndRestoresOnFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		context := NewWriteContext()
		restored, completed := false, false
		err := context.Run(func() error {
			require.True(t, context.IsScoped())
			leave, err := context.Enter(new(int), func() { restored = true })
			if err != nil {
				return err
			}
			defer leave()
			context.Defer(func() error { completed = true; return nil })
			if fail {
				return io.ErrClosedPipe
			}
			return nil
		})
		if fail {
			require.ErrorIs(t, err, io.ErrClosedPipe)
		} else {
			require.NoError(t, err)
		}
		require.True(t, restored)
		require.Equal(t, !fail, completed)
		require.False(t, context.IsScoped())
		_, err = context.Enter(new(int), func() {})
		require.ErrorIs(t, err, ErrWriteContextClosed)
	}
}

func TestWriteContextRunRejectsExistingOutstandingBindings(t *testing.T) {
	context := NewWriteContext()
	defer context.Close()
	leave, err := context.Enter(new(int), func() {})
	require.NoError(t, err)
	leave()
	called := false
	err = context.Run(func() error { called = true; return nil })
	require.ErrorIs(t, err, ErrWriteContextClosed)
	require.False(t, called)
	require.False(t, context.IsScoped())
	require.NoError(t, context.Finish())
}

func TestWriteContextRunRequiresANewContext(t *testing.T) {
	context := NewWriteContext()
	defer context.Close()
	context.Defer(func() error { return nil })
	require.ErrorIs(t, context.Run(func() error { return nil }), ErrWriteContextClosed)
	require.False(t, context.IsScoped())
}

func TestWriteContextSuspensionRejectsCompletionAndUnwindsNestedOwners(t *testing.T) {
	for _, outcome := range []string{"success", "error", "panic"} {
		t.Run(outcome, func(t *testing.T) {
			outer, middle, inner := NewWriteContext(), NewWriteContext(), NewWriteContext()
			completed, restored := 0, 0
			outer.Defer(func() error { completed++; return nil })
			assertSuspended := func(c *WriteContext) {
				require.ErrorIs(t, c.Finish(), ErrWriteContextSuspended)
				require.ErrorIs(t, c.Close(), ErrWriteContextSuspended)
				_, err := c.Enter(new(int), func() {})
				require.ErrorIs(t, err, ErrWriteContextSuspended)
			}
			run := func() error {
				return middle.Run(func() error {
					for range 2 {
						leave, err := middle.EnterFrom(new(int), outer, func() { restored++ })
						require.NoError(t, err)
						leave()
					}
					return inner.Run(func() error {
						for _, previous := range []*WriteContext{outer, middle} {
							leave, err := inner.EnterFrom(new(int), previous, func() { restored++ })
							require.NoError(t, err)
							leave()
						}
						assertSuspended(outer)
						assertSuspended(middle)
						require.Zero(t, completed)
						require.Zero(t, restored)
						switch outcome {
						case "error":
							return io.ErrClosedPipe
						case "panic":
							panic("write failed")
						}
						return nil
					})
				})
			}
			switch outcome {
			case "panic":
				require.PanicsWithValue(t, "write failed", func() { _ = run() })
			case "error":
				require.ErrorIs(t, run(), io.ErrClosedPipe)
			default:
				require.NoError(t, run())
			}
			require.Equal(t, 4, restored)
			require.NoError(t, outer.Finish())
			require.Equal(t, 1, completed)
			require.NoError(t, outer.Close())
		})
	}
}

func TestWriteContextRejectsClosedPreviousBindingWithoutAcquiringOwnership(t *testing.T) {
	previous, context := NewWriteContext(), NewWriteContext()
	require.NoError(t, previous.Close())
	restored := false
	err := context.Run(func() error {
		_, err := context.EnterFrom(new(int), previous, func() { restored = true })
		return err
	})
	require.ErrorIs(t, err, ErrWriteContextClosed)
	require.False(t, restored)
	require.Zero(t, previous.suspended)
}

func TestWriteContextSuspendsFinishedBindingsUntilRestorationCompletes(t *testing.T) {
	previous, context := NewWriteContext(), NewWriteContext()
	require.NoError(t, previous.Finish())
	err := context.Run(func() error {
		leave, err := context.EnterFrom(new(int), previous, func() {
			require.ErrorIs(t, previous.Close(), ErrWriteContextSuspended)
		})
		if err != nil {
			return err
		}
		leave()
		require.ErrorIs(t, previous.Close(), ErrWriteContextSuspended)
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, previous.Close())
}
