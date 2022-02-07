package harness

func EmptyTestSuiteResult() *testSuiteResult {
	return &testSuiteResult{
		Error: nil,
	}
}
