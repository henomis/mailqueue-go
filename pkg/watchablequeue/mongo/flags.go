package mongowatchablequeue

import "sync"

type mongoWatchableQueueFlag struct {
	sync.Mutex
	watched bool
}

func (f *mongoWatchableQueueFlag) SetWatched(value bool) {
	f.Lock()
	defer f.Unlock()
	f.watched = value
}

func (f *mongoWatchableQueueFlag) IsWatched() bool {
	f.Lock()
	defer f.Unlock()
	return f.watched
}
