package webhook

import (
	"encoding/json"
	"time"
)

// Job represents a webhook delivery job
type Job struct {
	ID          string                 `json:"id"`
	EventType   string                 `json:"event_type"`
	UserID      string                 `json:"user_id"`
	Data        map[string]interface{} `json:"data"`
	Attempts    int                    `json:"attempts"`
	Status      string                 `json:"status"` // pending, completed, failed
	CreatedAt   time.Time              `json:"created_at"`
	NotBefore   time.Time              `json:"not_before"`
	LastError   string                 `json:"last_error,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// JobStatus constants
const (
	JobStatusPending   = "pending"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
)

// ToJSON serializes the job to JSON
func (j *Job) ToJSON() ([]byte, error) {
	return json.Marshal(j)
}

// FromJSON deserializes a job from JSON
func FromJSON(data []byte) (*Job, error) {
	var job Job
	err := json.Unmarshal(data, &job)
	return &job, err
}
