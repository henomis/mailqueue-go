package limiter

import (
	"time"
)

//RealSleeper is a real sleeper
type RealSleeper struct {
}

//Now returns current time
func (s *RealSleeper) Now() time.Time {
	return time.Now()
}

//Sleep zZZzZ..
func (s *RealSleeper) Sleep(interval time.Duration) {
	time.Sleep(interval)
}
