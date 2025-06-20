//go:build examples

package malformed

import "testing"

// TestWithoutTParameter demonstrates a test function with wrong signature
func TestWithoutTParameter() {
	// This should cause go test to report issues
	panic("This test has wrong signature")
}

// TestWithWrongParameter demonstrates another malformed test
func TestWithWrongParameter(wrong int) {
	// This should also cause issues
}

// Not a test function but looks like one
func testNotExported(t *testing.T) {
	t.Log("This won't run as a test")
}
