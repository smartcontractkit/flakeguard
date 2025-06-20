//go:build examples

package missing_deps

import (
	"testing"
)

func TestMissingDependency(t *testing.T) {
	// This test imports a package that doesn't exist
	_ = nonexistent.SomeFunction()
	t.Log("This test should fail to build due to missing dependency")
}
