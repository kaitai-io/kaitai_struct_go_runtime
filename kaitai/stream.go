package kaitai

import (
	"io"
)

// APIVersion defines the currently used API version.
const APIVersion = 0x0001

// Difference: this stream is not autoload
type Stream struct {
	io.ReadSeeker
	io.Writer

	buf           [8]byte
	bitsLeft      int
	bits          uint64
	bitsLe        bool
	bitsWriteMode bool

	childStreams     []*Stream
	writebackHandler *WriteBackHandler
}

func NewStream(i io.ReadSeeker) *Stream {
	return &Stream{ReadSeeker: i}
}

func (k *Stream) ToByteArray() ([]byte, error) {
	pos, err := k.Pos()
	if err != nil {
		return nil, err
	}
	_, err = k.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	b, err := k.ReadBytesFull()
	if err != nil {
		return nil, err
	}
	_, err = k.Seek(pos, io.SeekStart)
	return b, err
}

func (k *Stream) WriteBackChildStreams() error {
	return k.writeBackChildStreams(nil)
}

func (k *Stream) writeBackChildStreams(parent *Stream) error {
	pos, err := k.Pos()
	if err != nil {
		return err
	}
	for _, child := range k.childStreams {
		err = child.writeBackChildStreams(k)
		if err != nil {
			return err
		}
	}
	k.childStreams = k.childStreams[:0]
	_, err = k.Seek(pos, io.SeekStart)
	if err != nil {
		return err
	}
	if parent != nil {
		err = k.writeback(parent)
	}

	return err
}

func (k *Stream) SetWriteBackHandler(handler *WriteBackHandler) {
	k.writebackHandler = handler
}

func (k *Stream) writeback(parent *Stream) error {
	return k.writebackHandler.writeBack(parent)
}

func (k *Stream) AddChildStream(child *Stream) {
	if child.Writer == nil {
		child.Writer = k.Writer
	}
	k.childStreams = append(k.childStreams, child)
}

type ReadWriteStream struct {
	Stream

	CheckFunc     func() error
	FetchInstance func() error
}

func NewReadWriteStream(s *Stream, w io.Writer) *ReadWriteStream {
	s.Writer = w

	return &ReadWriteStream{Stream: *s}
}

func (ss *ReadWriteStream) WriteStream() error {
	err := ss.WriteSeq()
	if err != nil {
		return err
	}
	err = ss.FetchInstance()
	if err != nil {
		return err
	}
	err = ss.WriteBackChildStreams()
	return err
}

func (ss *ReadWriteStream) WriteSeq() error {
	return ss.WriteBackChildStreams()
}

type WriteBackHandler struct {
	pos     int64
	handler func(*Stream) error
}

func NewWriteBackHandler(pos int64, handler func(*Stream) error) *WriteBackHandler {
	return &WriteBackHandler{
		pos, handler,
	}
}

func (h *WriteBackHandler) writeBack(parent *Stream) error {
	_, err := parent.Seek(h.pos, io.SeekStart)
	if err != nil {
		return err
	}
	return h.handler(parent)
}
