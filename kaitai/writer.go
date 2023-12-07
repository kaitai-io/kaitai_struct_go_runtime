package kaitai

import (
	"io"
)

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

type ReadWriteTrait interface {
	FetchInstances() error
	WriteSeq() error
	Check() error
}

type ReadWriteStream struct {
	Stream

	FetchInstance func() error
	WriteSeq      func() error
	Check         func() error
}

func NewReadWriteStream(stream *Stream) *ReadWriteStream {
	return &ReadWriteStream{
		Stream: *stream,
	}
}

func (k *ReadWriteStream) setFetchInstanceHandler(handler func() error) {
	k.FetchInstance = handler
}

func (k *ReadWriteStream) setWriteSeqHandler(handler func() error) {
	k.WriteSeq = handler
}

func (k *ReadWriteStream) setCheckHandler(handler func() error) {
	k.Check = handler
}

func (k *ReadWriteStream) Write(rt ReadWriteTrait) error {
	k.setFetchInstanceHandler(rt.FetchInstances)
	k.setWriteSeqHandler(rt.WriteSeq)
	k.setCheckHandler(rt.Check)

	err := k.Check()
	if err != nil {
		return err
	}
	err = k.WriteSeq()
	if err != nil {
		return err
	}
	err = k.FetchInstance()
	if err != nil {
		return err
	}
	err = k.WriteBackChildStreams()
	return err
}
