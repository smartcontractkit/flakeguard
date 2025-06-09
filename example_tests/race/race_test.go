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

func TestSomeRace(t *testing.T) {
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
