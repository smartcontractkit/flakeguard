//go:build examples

package broken

import "testing"

func TestBroken(t *testing.T) {
	t.Fatal("I shouldn't be able to compile")
	// This test is intentionally broken and unable to compile to help us test build errors.
	// The line below is intentionally broken Go:
	var v int = "not an int"
}
