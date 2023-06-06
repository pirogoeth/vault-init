package forwarder

import (
	"context"
	"io"
	"testing"
)

func TestForwarderv2Tee(t *testing.T) {
	testPayload := "Hello world!"
	ctx, cancel := context.WithCancel(context.Background())
	rChild, wChild := io.Pipe()
	rTee, wTee := io.Pipe()

	fwd := New(rChild)
	fwd.Tee(wTee)

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
