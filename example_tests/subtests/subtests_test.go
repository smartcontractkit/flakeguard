//go:build examples

package subtests

import (
	"fmt"
	"testing"
)

func TestPass(t *testing.T) {
	t.Run("passing subtest", func(t *testing.T) {
		t.Parallel()
		t.Log("passing subtest")
	})

	t.Run(fmt.Sprintf("passing subtest with dynamic name %d", 1), func(t *testing.T) {
		t.Parallel()
		t.Log("passing subtest with dynamic name")
	})
}

func TestFail(t *testing.T) {
	t.Run("failing subtest", func(t *testing.T) {
		t.Parallel()
		t.Log("failing subtest")
		t.Fail()
	})

	t.Run(fmt.Sprintf("failing subtest with dynamic name %d", 1), func(t *testing.T) {
		t.Parallel()
		t.Log("failing subtest with dynamic name")
		t.Fail()
	})
}
