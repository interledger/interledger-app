package jobs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobToJSON(t *testing.T) {
	now := time.Now().UTC()
	completedAt := now.Add(5 * time.Minute)

	job := &Job{
		ID:      "job-123",
		JobType: "deposit_webhook",
		Data: map[string]interface{}{
			"deposit_id": "dep-456",
			"amount":     "100.50",
		},
		Attempts:    2,
		Status:      JobStatusPending,
		CreatedAt:   now,
		NotBefore:   now.Add(2 * time.Minute),
		LastError:   "connection timeout",
		CompletedAt: &completedAt,
	}

	// Serialize to JSON
	jsonBytes, err := job.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, jsonBytes)

	// Verify JSON contains expected fields
	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, "job-123")
	assert.Contains(t, jsonStr, "deposit_webhook")
	assert.Contains(t, jsonStr, "dep-456")
	assert.Contains(t, jsonStr, "100.50")
	assert.Contains(t, jsonStr, "connection timeout")
}

func TestJobRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond) // Truncate for JSON precision

	originalJob := &Job{
		ID:      "job-789",
		JobType: "webhook_delivery",
		Data: map[string]interface{}{
			"url":    "https://example.com/webhook",
			"method": "POST",
		},
		Attempts:  5,
		Status:    JobStatusFailed,
		CreatedAt: now,
		NotBefore: now.Add(10 * time.Minute),
		LastError: "max attempts reached",
	}

	// Serialize
	jsonBytes, err := originalJob.ToJSON()
	require.NoError(t, err)

	// Deserialize
	deserializedJob, err := FromJSON(jsonBytes)
	require.NoError(t, err)
	assert.NotNil(t, deserializedJob)

	// Verify all fields match
	assert.Equal(t, originalJob.ID, deserializedJob.ID)
	assert.Equal(t, originalJob.JobType, deserializedJob.JobType)
	assert.Equal(t, originalJob.Attempts, deserializedJob.Attempts)
	assert.Equal(t, originalJob.Status, deserializedJob.Status)
	assert.Equal(t, originalJob.LastError, deserializedJob.LastError)
	assert.WithinDuration(t, originalJob.CreatedAt, deserializedJob.CreatedAt, time.Millisecond)
	assert.WithinDuration(t, originalJob.NotBefore, deserializedJob.NotBefore, time.Millisecond)
	assert.Equal(t, originalJob.Data["url"], deserializedJob.Data["url"])
	assert.Equal(t, originalJob.Data["method"], deserializedJob.Data["method"])
}

func TestFromJSONInvalidData(t *testing.T) {
	invalidJSON := []byte(`{invalid json}`)
	job, err := FromJSON(invalidJSON)
	assert.Error(t, err)
	assert.Nil(t, job)
}

func TestFromJSONEmptyData(t *testing.T) {
	emptyJSON := []byte(`{}`)
	job, err := FromJSON(emptyJSON)
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Empty(t, job.ID)
	assert.Empty(t, job.JobType)
}

func TestJobStatusConstants(t *testing.T) {
	// Verify status constants are defined correctly
	assert.Equal(t, "pending", JobStatusPending)
	assert.Equal(t, "completed", JobStatusCompleted)
	assert.Equal(t, "failed", JobStatusFailed)
}
