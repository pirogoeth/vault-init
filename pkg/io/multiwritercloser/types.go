package multiwritercloser

import (
	"io"
)

var _ io.Writer = (*MultiWriterCloser)(nil)
var _ io.Closer = (*MultiWriterCloser)(nil)

// MultiWriterCloser is a wrapper around io.MultiWriter that supports
// io.Writer and io.Closer impls.
type MultiWriterCloser struct {
	io.Writer

	cs []io.Closer
}
