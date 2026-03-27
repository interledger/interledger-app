package jobs

import "errors"

var ErrQueueClosed = errors.New("queue is closed")

// Queue accepts background jobs.
type Queue interface {
	Enqueue(job Job) error
	Close()
}

type InMemoryQueue struct {
	ch     chan Job
	closed chan struct{}
}

func NewInMemoryQueue(buffer int) *InMemoryQueue {
	if buffer < 1 {
		buffer = 16
	}

	return &InMemoryQueue{
		ch:     make(chan Job, buffer),
		closed: make(chan struct{}),
	}
}

func (q *InMemoryQueue) Enqueue(job Job) error {
	select {
	case <-q.closed:
		return ErrQueueClosed
	default:
	}

	select {
	case <-q.closed:
		return ErrQueueClosed
	case q.ch <- job:
		return nil
	}
}

func (q *InMemoryQueue) Close() {
	select {
	case <-q.closed:
		return
	default:
		close(q.closed)
	}
}

func (q *InMemoryQueue) Jobs() <-chan Job {
	return q.ch
}

type NoopQueue struct{}

func NewNoopQueue() *NoopQueue {
	return &NoopQueue{}
}

func (q *NoopQueue) Enqueue(_ Job) error { return nil }
func (q *NoopQueue) Close()              {}
