package watchablequeue

const (
	StatusEnqueued     = 200
	StatusDequeued     = 210
	StatusSending      = 220
	StatusSent         = 230
	StatusRead         = 240
	StatusErrorSending = 410
	StatusCanceled     = 420
)

type WatchableQueue interface {
	Enqueue(interface{}) error
	Dequeue(interface{}) error
	Watch(interface{}) (<-chan interface{}, error)
	Unwatch()
	Commit(interface{}) error
	SetStatus(interface{}, int64) error
	Get(interface{}) error
}
