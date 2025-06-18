//go:build examples

package skip

import "testing"

func TestSkip(t *testing.T) {
	t.Skip("I intentionally skip")
}
