package supervise

import (
	"context"
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
