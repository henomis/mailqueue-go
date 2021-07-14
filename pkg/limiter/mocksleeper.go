package limiter

import "time"

//MockSleeper emulate a clock
type MockSleeper struct {
	Ts time.Time
}

//Now returns current time
func (s *MockSleeper) Now() time.Time {
	return s.Ts
}

//Sleep ZzZzz..
func (s *MockSleeper) Sleep(interval time.Duration) {
	s.Ts = s.Ts.Add(interval)
}
