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
