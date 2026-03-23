package jobs

// Job status constants used by the queue and worker.
const (
	JobStatusQueued     = "queued"
	JobStatusProcessing = "processing"
	JobStatusDelivered  = "delivered"
	JobStatusFailed     = "failed"
)

// Job type constants for webhook delivery.
const (
	JobTypeUserAssessmentWebhook    = "user_assessment_webhook"
	JobTypeTransactionStatusWebhook = "transaction_status_webhook"
)
