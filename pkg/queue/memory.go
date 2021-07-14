package queue

import (
	"context"
	"errors"

	"github.com/henomis/mailqueue-go/pkg/email"
	"github.com/henomis/mailqueue-go/pkg/limiter"
)

//MemoryOptions for queue
type MemoryOptions struct {
	Size int
}

//Memory queue implementation
type Memory struct {
	Options MemoryOptions
	Limiter limiter.Limiter
	buffer  []*email.Email
	sp      int
}

//Attach memory queue
func (q *Memory) Attach(ctx context.Context) error {
	q.buffer = make([]*email.Email, q.Options.Size)
	q.sp = 0
	return nil
}

//Detach implementation in memory
func (q *Memory) Detach() error {
	q.buffer = []*email.Email{}
	return nil
}

//Enqueue implementation in memory
func (q *Memory) Enqueue(email *email.Email) error {

	q.buffer[q.sp%q.Options.Size] = email
	q.sp++
	if q.sp >= q.Options.Size {
		q.sp = 0
	}
	return nil
}

//Dequeue implementation in memory
func (q *Memory) Dequeue() (*email.Email, error) {

	if !q.Limiter.Allow() {
		return nil, errors.New(ErrLimitError)
	}

	return q.buffer[q.sp], nil
}

//Commit implementation in memory
func (q *Memory) Commit(email *email.Email) error {

	q.buffer[q.sp].Sent = true

	return nil
}
