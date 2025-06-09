//go:build examples

package fail

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestPass(t *testing.T) {
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

func TestFail(t *testing.T) {
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
