package assert

import (
	"testing"
)

func Equal[T comparable](t *testing.T, actual, expected T) {
	// Idicates that this is a test helper. So when t.Errorf gets called here,
	// the Go test runner will report the filename and line number of he code
	// which called this method.
	t.Helper()

	if actual != expected {
		t.Errorf("got %v; want %v", actual, expected)
	}
}
