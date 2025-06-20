//go:build examples

package fail

import (
	"testing"
)

func TestFail(t *testing.T) {
	t.Parallel()

	t.Log("I fail")
	t.FailNow()
}
