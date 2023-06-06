package logformatter

import "testing"

func TestConfigure(t *testing.T) {
	var err error

	_, err = Configure("default")
	if err != nil {
		t.Errorf("unexpected error configuring for 'default'")
	}

	_, err = Configure("plain")
	if err != nil {
		t.Errorf("unexpected error configuring for 'plain'")
	}

	_, err = Configure("json")
	if err != nil {
		t.Errorf("unexpected error configuring for 'json'")
	}

	_, err = Configure("noop")
	if err == nil {
		t.Errorf("expected error while configuring for 'noop'")
	}
}
