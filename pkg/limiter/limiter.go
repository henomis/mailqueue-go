package limiter

import "time"

//Limiter interface
type Limiter interface {
	Allow() bool
}

//Sleeper interface
type Sleeper interface {
	Now() time.Time
	Sleep(time.Duration)
}
