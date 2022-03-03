package forwarder

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"glow.dev.maio.me/seanj/vault-init/pkg/io/multiwritercloser"
)

func New(rc io.ReadCloser) *Forwarder {
	return &Forwarder{
		doneCh:  make(chan error),
		inlet:   rc,
		outlets: []io.WriteCloser{},
		ctx:     nil,
		cancel:  nil,
	}
}

// Tee adds an io.Writer to the list of outlets that will have messages
// written to them. Must be called before `Start` or will block.
func (f *Forwarder) Tee(w ...io.WriteCloser) {
	f.Lock()
	defer f.Unlock()

	f.outlets = append(f.outlets, w...)
}

func (f *Forwarder) Start(ctx context.Context) {
	// Intentionally leaving this forwarder locked. After closed or run,
	// a forwarder should no longer be used.
	f.Lock()

	f.ctx, f.cancel = context.WithCancel(ctx)

	go f.run()
}

func (f *Forwarder) Stop() {
	f.cancel()
}

// Close waits for the forwarder source to close, cancels the forwarder's context.
func (f *Forwarder) WaitClose() error {
outer:
	for {
		select {
		case err, ok := <-f.doneCh:
			if !ok {
				log.Debugf("Forwarder source channel closed: %s", err)
				break outer
			}
		case <-time.After(1 * time.Second):
			continue
		}
	}

	f.cancel()

	return nil
}

func (f *Forwarder) run() {
	multiOut := multiwritercloser.New(f.outlets...)
	log.Debugf("Forwarding to %#v", multiOut)

	if _, err := io.Copy(multiOut, f.inlet); err != nil {
		switch err := err.(type) {
		case *os.PathError:
			switch err.Err {
			case os.ErrClosed:
				f.doneCh <- fmt.Errorf("reader closed: %w", err.Err)
			}
		default:
			err = fmt.Errorf("unhandled error in forwarder: %w", err)
			f.doneCh <- err
		}
	}

	close(f.doneCh)
	multiOut.Close()
}
