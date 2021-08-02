package util

import (
	"bufio"
	"context"
	"io"
	"time"
)

func WaitUntilCompletion(ctx context.Context, cancel context.CancelFunc, reader io.Reader) error {
	lines := make(chan string)
	go readToEnd(reader, lines)
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				cancel()
			} else {
				log.Debugf("Read: %s", line)
			}
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			continue
		}
	}
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
