package util

import (
	"bufio"
	"context"
	"io"
	"time"
)

type ReaderCallback = func(context.Context, string)

// ReadLinesToCallback takes a context.Context, io.Reader, and callback. It reads the io.Reader until its end,
// writing all lines in the reader to the callback function. Each line emitted will trigger a new callback.
func ReadLinesToCallback(ctx context.Context, cancel context.CancelFunc, reader io.Reader, callback ReaderCallback) error {
	lines := make(chan string)
	go readToEnd(reader, lines)
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				cancel()
			} else {
				callback(ctx, line)
			}
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			continue
		}
	}
}

func WaitUntilCompletion(ctx context.Context, cancel context.CancelFunc, reader io.Reader) error {
	return ReadLinesToCallback(ctx, cancel, reader, func(_ context.Context, line string) {
		log.Debugf("Read: %s", line)
	})
}

func readToEnd(reader io.Reader, dest chan<- string) {
	buf := bufio.NewScanner(reader)
	for {
		if ok := buf.Scan(); !ok {
			close(dest)
			if err := buf.Err(); err != nil {
				panic(err)
			}

			return
		}

		dest <- buf.Text()
	}
}
