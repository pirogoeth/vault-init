package supervise

import (
	"context"
	"io"
	"strings"

	"github.com/mitchellh/go-linereader"
	"github.com/sirupsen/logrus"
)

// NewForwarder initializes a forwarder instance with the given pipe pair
func newForwarder(stdoutPipe, stderrPipe io.ReadCloser) *forwarderv1 {
	stdoutLog := log.WithField("stream", "stdout").WriterLevel(logrus.InfoLevel)
	stderrLog := log.WithField("stream", "stderr").WriterLevel(logrus.InfoLevel)
	return &forwarderv1{
		cancel:     nil,
		stdoutCh:   linereader.New(stdoutPipe),
		stderrCh:   linereader.New(stderrPipe),
		outWriters: []io.Writer{stdoutLog},
		errWriters: []io.Writer{stderrLog},
	}
}

// TeeStdout adds an io.Writer to the list of writers that will have stdout messages
// written to them. Must be called before `Start` or will block.
func (f *forwarderv1) TeeStdout(w ...io.Writer) {
	f.Lock()
	defer f.Unlock()

	f.outWriters = append(f.outWriters, w...)
}

// TeeStderr adds an io.Writer to the list of writers that will have stderr messages
// written to them. Must be called before `Start` or will block.
func (f *forwarderv1) TeeStderr(w ...io.Writer) {
	f.Lock()
	defer f.Unlock()

	f.errWriters = append(f.errWriters, w...)
}

func (f *forwarderv1) Start(ctx context.Context) {
	fwdCtx, cancel := context.WithCancel(ctx)
	f.cancel = cancel

	go f.run(fwdCtx)
}

func (f *forwarderv1) Stop() {
	f.cancel()

	f.Lock()
}

func (f *forwarderv1) run(ctx context.Context) {
	f.Lock()
	defer f.Unlock()

	multiOut := io.MultiWriter(f.outWriters...)
	log.Debugf("Forwarding stdout to %#v", multiOut)
	multiErr := io.MultiWriter(f.errWriters...)
	log.Debugf("Forwarding stderr to %#v", multiErr)

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
			multiOut.Write([]byte("\x0a"))
		case line := <-f.stderrCh.Ch:
			if strings.TrimSpace(line) == "" {
				continue
			}

			multiErr.Write([]byte(line))
			multiErr.Write([]byte("\x0a"))
		}
	}
}
