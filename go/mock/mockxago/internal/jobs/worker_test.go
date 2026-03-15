package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/mock/mockxago/internal/models"
	"gitlab.com/fynbos/mock/mockxago/internal/storage"
)

func TestWorker_ProcessJobWithoutHandlerMarksFailed(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)
	w := NewWorker(q)

	jobID, err := q.Enqueue("unknown", map[string]interface{}{"k": "v"}, time.Now().Add(-time.Second))
	require.NoError(t, err)

	job, err := q.store.GetJob(context.Background(), jobID)
	require.NoError(t, err)
	require.NotNil(t, job)

	w.processJob(job)

	updated, err := q.store.GetJob(context.Background(), jobID)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "pending", updated.Status)
	assert.Equal(t, 1, updated.Attempts)
	assert.True(t, updated.NotBefore.After(time.Now().Add(20*time.Second)))
}

func TestWorker_ProcessJobWithHandlerSuccessCompletes(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)
	w := NewWorker(q)

	w.RegisterHandler("ok", func(ctx context.Context, job *models.Job) error {
		return nil
	})

	jobID, err := q.Enqueue("ok", map[string]interface{}{"k": "v"}, time.Now().Add(-time.Second))
	require.NoError(t, err)

	job, err := q.store.GetJob(context.Background(), jobID)
	require.NoError(t, err)
	require.NotNil(t, job)

	w.processJob(job)

	updated, err := q.store.GetJob(context.Background(), jobID)
	require.NoError(t, err)
	assert.Equal(t, "completed", updated.Status)
	assert.NotNil(t, updated.CompletedAt)
}

func TestWorker_ProcessReadyJobsWithFailureReschedules(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)
	w := NewWorker(q)

	w.RegisterHandler("retry", func(ctx context.Context, job *models.Job) error {
		return errors.New("temporary")
	})

	jobID, err := q.Enqueue("retry", map[string]interface{}{}, time.Now().Add(-time.Second))
	require.NoError(t, err)

	w.processReadyJobs()

	updated, err := q.store.GetJob(context.Background(), jobID)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "pending", updated.Status)
	assert.Equal(t, 1, updated.Attempts)
	assert.True(t, updated.NotBefore.After(time.Now().Add(20*time.Second)))
}

func TestWorker_StartStopAndStartAsync(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)
	w := NewWorker(q)

	done := make(chan struct{})
	w.Stop()
	go func() {
		w.Start()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("worker did not stop promptly")
	}

	w2 := NewWorker(NewQueue(storage.NewMemoryStorage()))
	w2.StartAsync()
	w2.Stop()
	time.Sleep(20 * time.Millisecond)
}
