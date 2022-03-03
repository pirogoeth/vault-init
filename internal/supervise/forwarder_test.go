package supervise

import (
	"context"
	"io"
	"testing"
)

var _ io.ReadCloser = (*nilReader)(nil)

func newNilReader() *nilReader { return &nilReader{} }

type nilReader struct{}

func (nr *nilReader) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (nr *nilReader) Close() error { return nil }

func TestForwarderTeeStdout(t *testing.T) {
	testPayload := "Hello world!"
	ctx, cancel := context.WithCancel(context.Background())
	rChild, wChild := io.Pipe()
	rTee, wTee := io.Pipe()

	fwd := newForwarder(rChild, newNilReader())
	fwd.TeeStdout(wTee)

	if _, err := wChild.Write([]byte(testPayload)); err != nil {
		t.Errorf("error while writing test data to pipe: %s", err)
	}

	if err := wChild.Close(); err != nil {
		t.Errorf("error while closing test write pipe: %s", err)
	}

	go fwd.Start(ctx)

	expectedCount := len(testPayload)
	buf := make([]byte, expectedCount)

	count, err := rTee.Read(buf)
	if count != expectedCount {
		t.Errorf("read count mismatch: expected %d, got %d", expectedCount, count)
	} else if err != nil {
		t.Errorf("error reading from pipe: %s", err)
	}

	if string(buf) != testPayload {
		t.Errorf("read payload did not match test payload: `%s` != `%s`", string(buf), testPayload)
	}

	cancel()
}

func TestForwarderTeeStderr(t *testing.T) {
	testPayload := "Hello world!"
	ctx, cancel := context.WithCancel(context.Background())
	rChild, wChild := io.Pipe()
	rTee, wTee := io.Pipe()

	fwd := newForwarder(newNilReader(), rChild)
	fwd.TeeStderr(wTee)

	if _, err := wChild.Write([]byte(testPayload)); err != nil {
		t.Errorf("error while writing test data to pipe: %s", err)
	}

	if err := wChild.Close(); err != nil {
		t.Errorf("error while closing test write pipe: %s", err)
	}

	go fwd.Start(ctx)

	expectedCount := len(testPayload)
	buf := make([]byte, expectedCount)

	count, err := rTee.Read(buf)
	if count != expectedCount {
		t.Errorf("read count mismatch: expected %d, got %d", expectedCount, count)
	} else if err != nil {
		t.Errorf("error reading from pipe: %s", err)
	}

	if string(buf) != testPayload {
		t.Errorf("read payload did not match test payload: `%s` != `%s`", string(buf), testPayload)
	}

	cancel()
}
