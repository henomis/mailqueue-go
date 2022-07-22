package limiter

import (
	"time"
)

type FixedWindow struct {
	Allowed  uint64
	Interval time.Duration
	ticker   *time.Ticker
	bucket   chan struct{}
}

func NewFixedWindowLimiter(allow uint64, interval time.Duration) *FixedWindow {

	limiter := &FixedWindow{
		Allowed:  allow,
		Interval: interval,
	}

	limiter.run()

	return limiter
}

func (lb *FixedWindow) Wait() chan struct{} {
	return lb.bucket
}

func (lb *FixedWindow) run() {

	lb.ticker = time.NewTicker(lb.Interval)
	lb.bucket = make(chan struct{}, lb.Allowed)
	lb.refill()

	go func() {
		for range lb.ticker.C {
			lb.refill()
		}
	}()

}

func (lb *FixedWindow) refill() {
	for i := 0; i < int(lb.Allowed); i++ {
		select {
		case lb.bucket <- struct{}{}:
		default:
			return
		}
	}
}
