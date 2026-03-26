package jobs

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueueAndWorkerExecutesJob(t *testing.T) {
	q := NewInMemoryQueue(4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go StartWorker(ctx, q)

	var called int32
	err := q.Enqueue(Job{Delay: 5 * time.Millisecond, Run: func(context.Context) error {
		atomic.AddInt32(&called, 1)
		return nil
	}})
	if err != nil {
		t.Fatalf("Enqueue() unexpected error: %v", err)
	}

	time.Sleep(30 * time.Millisecond)
	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("job should have executed once")
	}
}

func TestQueueClosedRejectsEnqueue(t *testing.T) {
	q := NewInMemoryQueue(1)
	q.Close()
	if err := q.Enqueue(Job{}); err != ErrQueueClosed {
		t.Fatalf("expected ErrQueueClosed got %v", err)
	}
}
