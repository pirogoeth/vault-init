package multiwritercloser

import (
	"io"
)

func New(wcs ...io.WriteCloser) *MultiWriterCloser {
	ws := []io.Writer{}
	for _, w := range wcs {
		ws = append(ws, io.Writer(w))
	}
	mwc := &MultiWriterCloser{Writer: io.MultiWriter(ws...)}
	for _, w := range wcs {
		if wc, ok := w.(io.Closer); ok {
			mwc.cs = append(mwc.cs, io.Closer(wc))
		}
	}

	return mwc
}

func (mwc *MultiWriterCloser) Close() (err error) {
	for _, c := range mwc.cs {
		if closeErr := c.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}

	return err
}
