package limiter

import "time"

//DefaultLimiter is very similar to fixed window limiter but it has variable window start timestamp
type DefaultLimiter struct {
	Allowed   int
	Interval  time.Duration
	Sleeper   Sleeper
	timestamp time.Time
	count     int
}

//NewDefaultLimiter create a new instance
func NewDefaultLimiter(allow int, interval time.Duration, sleeper Sleeper) *DefaultLimiter {
	return &DefaultLimiter{
		Allowed:  allow,
		Interval: interval,
		Sleeper:  sleeper,
	}
}

//Allow require permission to perform an action under a certain limiter
func (l *DefaultLimiter) Allow() bool {

	var now time.Time

	now = l.Sleeper.Now().UTC()

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
