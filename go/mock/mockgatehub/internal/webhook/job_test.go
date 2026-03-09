package webhook

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToJSON_Roundtrip(t *testing.T) {
	original := &Job{
		ID:        "job-1",
		EventType: "core.deposit.completed",
		UserID:    "user-abc",
		Data:      map[string]interface{}{"amount": "100.00"},
		Attempts:  2,
		Status:    JobStatusPending,
		CreatedAt: time.Now().UTC().Truncate(time.Millisecond),
		NotBefore: time.Now().UTC().Add(5 * time.Second).Truncate(time.Millisecond),
	}

	data, err := original.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := FromJSON(data)
	require.NoError(t, err)
	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.EventType, restored.EventType)
	assert.Equal(t, original.UserID, restored.UserID)
	assert.Equal(t, original.Attempts, restored.Attempts)
	assert.Equal(t, original.Status, restored.Status)
	assert.Equal(t, original.CreatedAt.Unix(), restored.CreatedAt.Unix())
}

func TestFromJSON_InvalidData(t *testing.T) {
	_, err := FromJSON([]byte("not json"))
	assert.Error(t, err)
}

func TestFromJSON_EmptyJob(t *testing.T) {
	job, err := FromJSON([]byte(`{}`))
	require.NoError(t, err)
	assert.Empty(t, job.ID)
	assert.Empty(t, job.EventType)
	assert.Equal(t, 0, job.Attempts)
}

func TestToJSON_WithCompletedAt(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	job := &Job{
		ID:          "job-done",
		Status:      JobStatusCompleted,
		CompletedAt: &now,
	}

	data, err := job.ToJSON()
	require.NoError(t, err)

	restored, err := FromJSON(data)
	require.NoError(t, err)
	require.NotNil(t, restored.CompletedAt)
	assert.Equal(t, now.Unix(), restored.CompletedAt.Unix())
}

func TestJobStatusConstants(t *testing.T) {
	assert.Equal(t, "pending", JobStatusPending)
	assert.Equal(t, "completed", JobStatusCompleted)
	assert.Equal(t, "failed", JobStatusFailed)
}
