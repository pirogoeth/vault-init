package supervise

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/mitchellh/go-linereader"
	log "github.com/sirupsen/logrus"
)

// NewForwarder initializes a forwarder instance with the given pipe pair
func newForwarder(stdoutPipe, stderrPipe io.ReadCloser) *forwarder {
	return &forwarder{
		cancel:    nil,
		stdoutCh:  linereader.New(stdoutPipe),
		stderrCh:  linereader.New(stderrPipe),
		outWriter: os.Stdout,
		errWriter: os.Stderr,
	}
}

func (f *forwarder) Start(ctx context.Context) {
	fwdCtx, cancel := context.WithCancel(ctx)
	f.cancel = cancel

	go f.run(fwdCtx)
}

func (f *forwarder) Stop() {
	f.cancel()
}

func (f *forwarder) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Infof("Child output forwarder exiting")
			return
		case line := <-f.stdoutCh.Ch:
			if strings.TrimSpace(line) == "" {
				continue
			}

			log.WithField("stream", "stdout").Infof(line)
		case line := <-f.stderrCh.Ch:
			if strings.TrimSpace(line) == "" {
				continue
			}

			log.WithField("stream", "stderr").Infof(line)
		}
	}
}
