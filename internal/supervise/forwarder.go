package supervise

import (
	"context"
	"io"
	"strings"

	"github.com/mitchellh/go-linereader"
	"github.com/sirupsen/logrus"
)

// NewForwarder initializes a forwarder instance with the given pipe pair
func newForwarder(stdoutPipe, stderrPipe io.ReadCloser) *forwarder {
	stdoutLog := log.WithField("stream", "stdout").WriterLevel(logrus.InfoLevel)
	stderrLog := log.WithField("stream", "stderr").WriterLevel(logrus.InfoLevel)
	return &forwarder{
		cancel:     nil,
		stdoutCh:   linereader.New(stdoutPipe),
		stderrCh:   linereader.New(stderrPipe),
		outWriters: []io.Writer{stdoutLog},
		errWriters: []io.Writer{stderrLog},
	}
}

// TeeStdout adds an io.Writer to the list of writers that will have stdout messages
// written to them. Must be called before `Start` or will block.
func (f *forwarder) TeeStdout(w ...io.Writer) {
	f.Lock()
	defer f.Unlock()

	f.outWriters = append(f.outWriters, w...)
}

// TeeStderr adds an io.Writer to the list of writers that will have stderr messages
// written to them. Must be called before `Start` or will block.
func (f *forwarder) TeeStderr(w ...io.Writer) {
	f.Lock()
	defer f.Unlock()

	f.errWriters = append(f.errWriters, w...)
}

func (f *forwarder) Start(ctx context.Context) {
	fwdCtx, cancel := context.WithCancel(ctx)
	f.cancel = cancel

	go f.run(fwdCtx)
}

func (f *forwarder) Stop() {
	f.cancel()

	f.Lock()
	defer f.Unlock()
}

func (f *forwarder) run(ctx context.Context) {
	f.Lock()
	defer f.Unlock()

	multiOut := io.MultiWriter(f.outWriters...)
	multiErr := io.MultiWriter(f.errWriters...)

outer:
	for {
		select {
		case <-ctx.Done():
			log.Infof("Child output forwarder exiting")
			break outer
		case line := <-f.stdoutCh.Ch:
			if strings.TrimSpace(line) == "" {
				continue
			}

			multiOut.Write([]byte(line))
		case line := <-f.stderrCh.Ch:
			if strings.TrimSpace(line) == "" {
				continue
			}

			multiErr.Write([]byte(line))
		}
	}
}
