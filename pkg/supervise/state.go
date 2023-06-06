package supervise

import (
	"context"
	"io"
)

func newState(parentCtx context.Context) *state {
	childCtx, childCancel := context.WithCancel(parentCtx)
	return &state{
		child:       nil,
		childCtx:    childCtx,
		childCancel: childCancel,
		parentCtx:   parentCtx,
	}
}

// replaceChild cancels the previous child's context, clears the `child`
// field, and creates a new child context and cancel func. Returns the old
// child.
func (s *state) replaceChildContext() {
	s.childCancel()

	newCtx, newCancel := context.WithCancel(s.parentCtx)
	s.childCtx = newCtx
	s.childCancel = newCancel
}

func (s *state) closeChildOutputs() error {
	var lastErr error

	if c, ok := s.child.Stdout.(io.Closer); ok {
		lastErr = c.Close()
	}

	if c, ok := s.child.Stderr.(io.Closer); ok {
		lastErr = c.Close()
	}

	return lastErr
}
