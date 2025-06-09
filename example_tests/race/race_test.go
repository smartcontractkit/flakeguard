//go:build examples

package race

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

var raceValue = ""

func TestRace(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}
	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			raceValue = fmt.Sprintf("race %d", i)
		}()
	}

	wg.Wait()
}

func TestPass(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		t.Run(fmt.Sprintf("pass-%d", i), func(t *testing.T) {
			// Sleep to make sure the race happens in the middle of test executions
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
			// Sleep to make sure the race happens in the middle of test executions
			sleep := time.Duration(rand.Intn(100)) * time.Millisecond
			t.Logf("Fail: Sleeping for %s", sleep)
			time.Sleep(sleep)
			t.Log("I fail")
			t.FailNow()
		})
	}
}
