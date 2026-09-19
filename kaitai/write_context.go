package kaitai

import "errors"

// ErrWriteContextClosed reports a missing, completed, or closed output session.
var ErrWriteContextClosed = errors.New("write context is missing or complete")

// WriteContext coordinates one generated output operation. It records only
// objects and child reservations participating in that operation, and completes
// pending instances in nested registration order before propagating children.
// Like Writer, it is not safe for concurrent use. Always defer Close after New.
type WriteContext struct {
	root      writeFrame
	current   *writeFrame
	objects   map[any]bool
	restore   []func()
	children  []*ChildWriter
	finishing bool
	finished  bool
	closed    bool
	scoped    bool
}

type writeFrame struct {
	action   func() error
	children []*writeFrame
}

// NewWriteContext starts an independent generated write.
func NewWriteContext() *WriteContext {
	c := &WriteContext{objects: make(map[any]bool)}
	c.current = &c.root
	return c
}

// Run owns an entire independent write: it invokes write, finishes pending
// operations if writing succeeds, and always closes the context before returning.
// The context must be new. Generated independent writers use this bounded scope
// to permit synchronous reentry of a graph already bound to a suspended session.
func (c *WriteContext) Run(write func() error) error {
	if c == nil || c.closed || c.finished || c.finishing || c.scoped ||
		len(c.objects) != 0 || len(c.root.children) != 0 || len(c.children) != 0 {
		return ErrWriteContextClosed
	}
	c.scoped = true
	defer c.Close()
	err := write()
	if err != nil {
		return err
	}
	return c.Finish()
}

// IsScoped reports whether Run currently owns this context's complete lifetime.
// A scoped context is guaranteed to restore its graph bindings before Run returns;
// a caller-managed context may instead remain outstanding after a sequence call.
func (c *WriteContext) IsScoped() bool {
	return c != nil && c.scoped && !c.closed
}

// Enter binds a generated object until Close. object must be a comparable
// identity (generated callers pass a pointer). The returned function ends its
// sequence registration scope; it does not release its output binding. A second
// outstanding occurrence of the same object is rejected rather than selecting
// an ambiguous stream. restore is run on Close even if writing fails.
func (c *WriteContext) Enter(object any, restore func()) (func(), error) {
	if c == nil || c.closed || c.finished {
		return nil, ErrWriteContextClosed
	}
	if c.objects[object] {
		return nil, NewConsistencyError("output binding", "one occurrence per object", "repeated object")
	}
	c.objects[object] = true
	c.restore = append(c.restore, restore)
	parent := c.current
	f := &writeFrame{}
	parent.children = append(parent.children, f)
	c.current = f
	return func() { c.current = parent }, nil
}

// Defer records one pending instance operation. Children registered by this
// operation finish before the next operation in the containing frame.
func (c *WriteContext) Defer(action func() error) {
	c.current.children = append(c.current.children, &writeFrame{action: action})
}

// RegisterChild records a reservation owned by this operation. Historical
// reservations in the same external writer are not part of the operation.
func (c *WriteContext) RegisterChild(child *ChildWriter) {
	c.children = append(c.children, child)
}

// Finish completes pending instances, then propagates reserved children without
// aligning their parents' pending partial bytes. Finish may be called only once;
// a failed generated write requires a fresh context and output. Explicit
// ChildWriter handles retain their separate WriteBack retry contract.
func (c *WriteContext) Finish() error {
	if c == nil || c.closed || c.finished || c.finishing {
		return ErrWriteContextClosed
	}
	c.finishing = true
	defer func() { c.finishing = false; c.finished = true }()
	err := c.finishFrame(&c.root)
	if err != nil {
		return err
	}
	for _, child := range c.children {
		if !child.flushed {
			err := child.WriteBackNoAlign()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Close restores all generated bindings in reverse entry order without writing
// output. It detaches this operation's child registrations on success or failure
// so a subsequent independent operation cannot flush abandoned reservations.
// Explicit child handles remain usable for direct runtime retry.
func (c *WriteContext) Close() {
	if c == nil || c.closed {
		return
	}
	c.closed = true
	for i := len(c.restore) - 1; i >= 0; i-- {
		c.restore[i]()
	}
	c.detachChildren()
	c.root = writeFrame{}
	c.current = nil
	c.objects = nil
	c.restore = nil
	c.children = nil
}

func (c *WriteContext) detachChildren() {
	owned := make(map[*ChildWriter]bool, len(c.children))
	parents := make(map[*Writer]bool)
	for _, child := range c.children {
		owned[child] = true
		parents[child.parent] = true
	}
	for parent := range parents {
		remaining := parent.children[:0]
		for _, child := range parent.children {
			if !owned[child] {
				remaining = append(remaining, child)
			}
		}
		clear(parent.children[len(remaining):])
		parent.children = remaining
	}
}

func (c *WriteContext) finishFrame(f *writeFrame) error {
	previous := c.current
	c.current = f
	defer func() { c.current = previous }()
	if f.action != nil {
		err := f.action()
		if err != nil {
			return err
		}
	}
	for _, child := range f.children {
		err := c.finishFrame(child)
		if err != nil {
			return err
		}
	}
	return nil
}
