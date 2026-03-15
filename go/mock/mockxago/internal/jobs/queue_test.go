package jobs

import (
	"context"
	"testing"
	"time"

	"gitlab.com/fynbos/mock/mockxago/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQueue(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)
	assert.NotNil(t, q)
	assert.Equal(t, store, q.store)
}

func TestQueue_Enqueue(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)

	jobType := "test.webhook"
	data := map[string]interface{}{"key": "value"}
	readyAt := time.Now()

	jobID, err := q.Enqueue(jobType, data, readyAt)
	require.NoError(t, err)
	assert.NotEmpty(t, jobID)

	// Verify job was saved
	ctx := context.Background()
	job, err := store.GetJob(ctx, jobID)
	require.NoError(t, err)
	require.NotNil(t, job)
	assert.Equal(t, jobType, job.JobType)
	assert.Equal(t, data, job.Data)
	assert.Equal(t, 0, job.Attempts)
	assert.Equal(t, "pending", job.Status)
}

func TestQueue_GetReadyJobs(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)

	// Enqueue jobs with different ready times
	now := time.Now()
	_, err := q.Enqueue("job1", map[string]interface{}{"n": 1}, now.Add(-1*time.Hour)) // Ready
	require.NoError(t, err)
	_, err = q.Enqueue("job2", map[string]interface{}{"n": 2}, now.Add(-30*time.Minute)) // Ready
	require.NoError(t, err)
	_, err = q.Enqueue("job3", map[string]interface{}{"n": 3}, now.Add(1*time.Hour)) // Not ready
	require.NoError(t, err)

	// Get ready jobs
	jobs, err := q.GetReadyJobs(10)
	require.NoError(t, err)

	// Should get 2 jobs (job1 and job2), sorted by NotBefore
	assert.Len(t, jobs, 2)
	assert.Equal(t, "job1", jobs[0].JobType)
	assert.Equal(t, "job2", jobs[1].JobType)
}

func TestQueue_GetReadyJobs_Limit(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)

	// Enqueue 5 ready jobs
	now := time.Now()
	for i := 0; i < 5; i++ {
		_, err := q.Enqueue("job", map[string]interface{}{"n": i}, now.Add(-1*time.Minute))
		require.NoError(t, err)
	}

	// Get only 3
	jobs, err := q.GetReadyJobs(3)
	require.NoError(t, err)
	assert.Len(t, jobs, 3)
}

func TestQueue_MarkCompleted(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)

	// Enqueue a job
	jobID, err := q.Enqueue("test.job", map[string]interface{}{}, time.Now())
	require.NoError(t, err)

	// Mark it completed
	err = q.MarkCompleted(jobID)
	require.NoError(t, err)

	// Verify status
	ctx := context.Background()
	job, err := store.GetJob(ctx, jobID)
	require.NoError(t, err)
	assert.Equal(t, "completed", job.Status)
	assert.NotNil(t, job.CompletedAt)
}

func TestQueue_MarkFailed_Retry(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)

	// Enqueue a job
	jobID, err := q.Enqueue("test.job", map[string]interface{}{}, time.Now())
	require.NoError(t, err)

	// Mark it failed (should reschedule)
	errMsg := "temporary error"
	err = q.MarkFailed(jobID, errMsg)
	require.NoError(t, err)

	// Verify attempts incremented and job rescheduled
	ctx := context.Background()
	job, err := store.GetJob(ctx, jobID)
	require.NoError(t, err)
	assert.Equal(t, 1, job.Attempts)
	assert.Equal(t, "pending", job.Status)
	assert.True(t, job.NotBefore.After(time.Now().Add(25*time.Second))) // Should be ~30s in future
}

func TestQueue_MarkFailed_MaxAttempts(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)

	// Enqueue a job
	jobID, err := q.Enqueue("test.job", map[string]interface{}{}, time.Now())
	require.NoError(t, err)

	// Fail it 10 times (maxAttempts)
	for i := 0; i < 10; i++ {
		err = q.MarkFailed(jobID, "error")
		require.NoError(t, err)
	}

	// Verify it's permanently failed
	ctx := context.Background()
	job, err := store.GetJob(ctx, jobID)
	require.NoError(t, err)
	assert.Equal(t, 10, job.Attempts)
	assert.Equal(t, "failed", job.Status)
	assert.NotNil(t, job.CompletedAt)
}

func TestQueue_PendingCount(t *testing.T) {
	store := storage.NewMemoryStorage()
	q := NewQueue(store)

	// Initially zero
	count := q.PendingCount()
	assert.Equal(t, 0, count)

	// Add 3 pending jobs
	now := time.Now()
	for i := 0; i < 3; i++ {
		_, err := q.Enqueue("job", map[string]interface{}{}, now.Add(-1*time.Minute))
		require.NoError(t, err)
	}

	// Add 1 future job (not ready)
	_, err := q.Enqueue("future", map[string]interface{}{}, now.Add(1*time.Hour))
	require.NoError(t, err)

	// Should count 3 ready jobs
	count = q.PendingCount()
	assert.Equal(t, 3, count)
}
