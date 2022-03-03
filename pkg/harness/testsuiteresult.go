package harness

import "io"

func newTestSuiteResult(rStderr, rStdout io.ReadCloser) *testSuiteResult {
	return &testSuiteResult{
		Error:        nil,
		StderrReader: rStderr,
		StdoutReader: rStdout,
	}
}
