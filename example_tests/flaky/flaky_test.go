//go:build examples

package flaky

import (
	"math/rand"
	"testing"
)

func TestNoFlake(t *testing.T) {
	t.Parallel()
	t.Log("I don't flake")
}

func TestFlakeTenPercent(t *testing.T) {
	t.Parallel()

	if rand.Intn(10) == 0 {
		t.Log("I flake 10% of the time")
		t.FailNow()
	}
}

func TestFlakeTwentyFivePercent(t *testing.T) {
	t.Parallel()

	if rand.Intn(4) == 0 {
		t.Log("I flake 25% of the time")
		t.FailNow()
	}
}

func TestFlakeFiftyPercent(t *testing.T) {
	t.Parallel()

	if rand.Intn(2) == 0 {
		t.Log("I flake 50% of the time")
		t.FailNow()
	}
}

func TestFlakeHundredPercent(t *testing.T) {
	t.Parallel()

	t.Log("I flake 100% of the time")
	t.FailNow()
}
