package limiter

import (
	"testing"
	"time"
)

var allowedRequests = 30
var allowedInterval = 1 * time.Minute

func TestAllow(t *testing.T) {

	s := &MockSleeper{
		Ts: time.Now(),
	}

	l := &DefaultLimiter{
		Allowed:  allowedRequests,
		Interval: allowedInterval,
		Sleeper:  s,
	}

	allowed := 0
	rejected := 0

	t.Run("100 requests", func(t *testing.T) {

		t.Helper()

		for i := 0; i < 100; i++ {
			if l.Allow() {
				allowed++
			} else {
				rejected++
			}
		}

		if allowed != allowedRequests {
			t.Errorf("Expected %d got %d", allowedRequests, allowed)
		}
	})

	t.Run("100 sleeping requests", func(t *testing.T) {

		t.Helper()

		for i := 0; i < 100; i++ {
			if l.Allow() {
				allowed++
			} else {
				rejected++
			}

			s.Sleep(1 * time.Second)
		}

		expected := int(100/allowed) * allowed

		if allowed != expected {
			t.Errorf("Expected %d got %d", expected, allowed)
		}
	})

}
