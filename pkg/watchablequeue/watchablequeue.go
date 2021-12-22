package watchablequeue

type WatchableQueue interface {
	Enqueue(interface{}) error
	Dequeue(interface{}) error
	Watch(interface{}) (<-chan interface{}, error)
	Commit(interface{}) error
	Unwatch()
}
