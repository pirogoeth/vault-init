package supervise

import (
	"context"
)

func newEvent(parentCtx context.Context) *event {
	childCtx, childCancel := context.WithCancel(parentCtx)
	return &event{
		child:       nil,
		childCtx:    childCtx,
		childCancel: childCancel,
		parentCtx:   parentCtx,
	}
}

// replaceChild cancels the previous child's context, clears the `child`
// field, and creates a new child context and cancel func. Returns the old
// child.
func (e *event) replaceChildContext() {
	e.childCancel()

	newCtx, newCancel := context.WithCancel(e.parentCtx)
	e.childCtx = newCtx
	e.childCancel = newCancel
}
