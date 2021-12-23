package limiter

import (
	"testing"
	"time"
)

var allowedRequests = int64(30)
var allowedInterval = 1 * time.Minute

func TestAllow(t *testing.T) {

	l := &DefaultLimiter{
		Allowed:  allowedRequests,
		Interval: allowedInterval,
	}

	allowed := int64(0)
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

			time.Sleep(1 * time.Second)
		}

		expected := int64(100/allowed) * allowed

		if allowed != expected {
			t.Errorf("Expected %d got %d", expected, allowed)
		}
	})

}
