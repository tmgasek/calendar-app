package assert

import (
	"strings"
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

func StringContains(t *testing.T, s, substr string) {
	t.Helper()

	if !strings.Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}

// NilError checks if the actual error is nil.
func NilError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Errorf("got: %v; expected: nil", actual)
	}
}

func Greater(t *testing.T, actual, expected int) {
	t.Helper()

	if actual <= expected {
		t.Errorf("got: %d; expected: greater than %d", actual, expected)
	}
}

func NotNil(t *testing.T, actual interface{}) {
	t.Helper()

	if actual == nil {
		t.Errorf("got: nil; expected: not nil")
	}
}
