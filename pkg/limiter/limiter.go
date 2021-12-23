package limiter

//Limiter interface
type Limiter interface {
	Allow() bool
}
