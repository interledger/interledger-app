package v1

const (
	contentTypeHeader = "Content-Type"
	contentTypeValue  = "application/json"

	acceptHeader = "Accept"
	acceptValue  = contentTypeValue

	clientIDHeader  = "x-pti-client-id"
	signatureHeader = "x-pti-signature"

	ptiRequestIDHeader  = "x-pti-request-id"
	ptiScenarioIDHeader = "x-pti-scenario-id"

	// for later updates
	// ptiDisableWebhookHeader = "x-pti-disable-webhook"
	// ptiSessionIDHeader      = "x-pti-session-id"
)
