package jobs

import (
	"encoding/json"
	"time"
)

// Job status constants
const (
	JobStatusPending   = "pending"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
)

// Job represents a unit of async work (deposit processing, webhook delivery, etc.)
type Job struct {
	ID          string                 `json:"id"`
	JobType     string                 `json:"job_type"`
	Data        map[string]interface{} `json:"data"`
	Attempts    int                    `json:"attempts"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	NotBefore   time.Time              `json:"not_before"`
	LastError   string                 `json:"last_error"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// ToJSON serializes the job to JSON bytes
func (j *Job) ToJSON() ([]byte, error) {
	return json.Marshal(j)
}

// FromJSON deserializes a job from JSON bytes
func FromJSON(data []byte) (*Job, error) {
	var j Job
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, err
	}
	return &j, nil
}
