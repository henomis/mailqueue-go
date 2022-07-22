package limiter

//Limiter interface
type Limiter interface {
	Wait() chan struct{}
}
