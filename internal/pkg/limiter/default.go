package limiter

import "time"

//DefaultLimiter is very similar to fixed window limiter but it has variable window start timestamp
type DefaultLimiter struct {
	Allowed   int64
	Interval  time.Duration
	timestamp time.Time
	count     int64
}

//NewDefaultLimiter create a new instance
func NewDefaultLimiter(allow int64, interval time.Duration) *DefaultLimiter {
	return &DefaultLimiter{
		Allowed:  allow,
		Interval: interval,
	}
}

//Allow require permission to perform an action under a certain limiter
func (l *DefaultLimiter) Allow() bool {

	now := time.Now().UTC()

	if now.Sub(l.timestamp) >= l.Interval {
		l.count = 0
		l.timestamp = now
	}

	if l.count >= l.Allowed {
		return false
	}

	l.count++
	return true

}
