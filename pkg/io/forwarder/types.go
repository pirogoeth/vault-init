package forwarder

import (
	"context"
	"io"
	"sync"
)

// Forwarder takes a io.ReadCloser from a child program
// and muxes it into any number of io.WriteCloser "outlets"
type Forwarder struct {
	sync.Mutex

	doneCh  chan error
	inlet   io.Reader
	outlets []io.WriteCloser

	ctx    context.Context
	cancel context.CancelFunc
}
