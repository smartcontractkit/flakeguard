package nested

import (
	"testing"
)

func TestNestedFail(t *testing.T) {
	t.Parallel()
	t.Log("I'm a test in a nested go.mod, and I'm going to fail")
	t.Fail()
}

func TestNestedPass(t *testing.T) {
	t.Parallel()
	t.Log("I'm a test in a nested go.mod, and I'm going to pass")
}
