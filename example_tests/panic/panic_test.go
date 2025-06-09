//go:build examples

package panic

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestPanic(t *testing.T) {
	t.Parallel()

	sleep := time.Duration(rand.Intn(1000))*time.Millisecond + 100*time.Millisecond
	t.Logf("Panic:Sleeping for %s", sleep)
	time.Sleep(sleep)

	panic(fmt.Sprintf("I slept for %s and panicked", sleep))
}

// nolint:paralleltest
func TestPasses(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		t.Run(fmt.Sprintf("pass-%d", i), func(t *testing.T) {
			sleep := time.Duration(rand.Intn(100)) * time.Millisecond
			t.Logf("Pass: Sleeping for %s", sleep)
			time.Sleep(sleep)
			t.Log("I pass")
		})
	}
}

// nolint:paralleltest
func TestFails(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		t.Run(fmt.Sprintf("fail-%d", i), func(t *testing.T) {
			sleep := time.Duration(rand.Intn(100)) * time.Millisecond
			t.Logf("Fail: Sleeping for %s", sleep)
			time.Sleep(sleep)
			t.Log("I fail")
			t.FailNow()
		})
	}
}
