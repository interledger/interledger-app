package webhook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
	"gitlab.com/fynbos/mock/mockgatehub/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	manager := NewManager("http://example.com/webhook", "secret", nil, nil, "")

	assert.NotNil(t, manager)
	assert.Equal(t, "http://example.com/webhook", manager.webhookURL)
	assert.Equal(t, "secret", manager.webhookSecret)
	assert.NotNil(t, manager.httpClient)
}

func TestSendAsync_NoURL(t *testing.T) {
	manager := NewManager("", "secret", nil, nil, "")

	// Should not panic when queue is nil and URL is empty
	manager.SendAsync("test.event", "user-123", map[string]interface{}{}, 0)
}

func TestHasURL_WithGlobalURL(t *testing.T) {
	m := NewManager("http://example.com/webhook", "secret", nil, nil, "")
	assert.True(t, m.HasURL())
}

func TestHasURL_NoURL(t *testing.T) {
	m := NewManager("", "secret", nil, nil, "")
	assert.False(t, m.HasURL())
}

func TestHasURL_NilManager(t *testing.T) {
	var m *Manager
	assert.False(t, m.HasURL())
}

func TestHasURL_FromOrgConfig(t *testing.T) {
	store := storage.NewMemoryStorage()
	_ = store.CreateOrganization(&models.Organization{
		ID:         "test-org",
		APIBaseURL: "http://org-callback.example.com",
	})
	m := NewManager("", "secret", nil, store, "test-org")
	assert.True(t, m.HasURL())
}

func TestResolveCallbackURL_OrgTakesPriority(t *testing.T) {
	store := storage.NewMemoryStorage()
	_ = store.CreateOrganization(&models.Organization{
		ID:         "org-1",
		APIBaseURL: "http://org-url.example.com",
	})
	m := NewManager("http://global.example.com", "secret", nil, store, "org-1")
	assert.Equal(t, "http://org-url.example.com", m.ResolveCallbackURL())
}

func TestResolveCallbackURL_FallsBackToGlobal(t *testing.T) {
	store := storage.NewMemoryStorage()
	// no org config
	m := NewManager("http://global.example.com", "secret", nil, store, "missing-org")
	assert.Equal(t, "http://global.example.com", m.ResolveCallbackURL())
}

func TestResolveOrgBaseURL_WithOrg(t *testing.T) {
	store := storage.NewMemoryStorage()
	_ = store.CreateOrganization(&models.Organization{
		ID:         "org-1",
		APIBaseURL: "http://org.example.com/",
	})
	m := NewManager("", "secret", nil, store, "org-1")
	assert.Equal(t, "http://org.example.com", m.ResolveOrgBaseURL()) // trailing slash trimmed
}

func TestResolveOrgBaseURL_NoOrg(t *testing.T) {
	store := storage.NewMemoryStorage()
	m := NewManager("http://global.example.com", "secret", nil, store, "nope")
	assert.Equal(t, "", m.ResolveOrgBaseURL())
}

func TestResolveOrgBaseURL_NoStore(t *testing.T) {
	m := NewManager("", "secret", nil, nil, "")
	assert.Equal(t, "", m.ResolveOrgBaseURL())
}

// --- normalizeVerificationPayload ---

func TestNormalizeVerificationPayload_AddsDefaults(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	result := normalizeVerificationPayload("id.verification.accepted", data)
	assert.Equal(t, "paywiser", result["gateway"])
	verified := result["verified"].(map[string]interface{})
	assert.Equal(t, "accepted", verified["short"])
	assert.Equal(t, 1, verified["status"])
}

func TestNormalizeVerificationPayload_Rejected(t *testing.T) {
	data := map[string]interface{}{}
	result := normalizeVerificationPayload("id.verification.rejected", data)
	verified := result["verified"].(map[string]interface{})
	assert.Equal(t, "rejected", verified["short"])
	assert.Equal(t, 2, verified["status"])
}

func TestNormalizeVerificationPayload_ActionRequired(t *testing.T) {
	data := map[string]interface{}{}
	result := normalizeVerificationPayload("id.verification.action_required", data)
	verified := result["verified"].(map[string]interface{})
	assert.Equal(t, "action_required", verified["short"])
	assert.Equal(t, 0, verified["status"])
}

func TestNormalizeVerificationPayload_NonVerificationEvent(t *testing.T) {
	data := map[string]interface{}{"amount": "100"}
	result := normalizeVerificationPayload("core.deposit.completed", data)
	assert.Equal(t, "100", result["amount"])
	_, hasGateway := result["gateway"]
	assert.False(t, hasGateway)
}

func TestNormalizeVerificationPayload_PreservesExistingFields(t *testing.T) {
	data := map[string]interface{}{
		"gateway": "custom",
		"verified": map[string]interface{}{
			"short":  "custom",
			"status": 99,
		},
	}
	result := normalizeVerificationPayload("id.verification.accepted", data)
	assert.Equal(t, "custom", result["gateway"]) // not overwritten
}

// --- coerceToMap ---

func TestCoerceToMap_Nil(t *testing.T) {
	result := coerceToMap(nil)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestCoerceToMap_Map(t *testing.T) {
	m := map[string]interface{}{"key": "value"}
	result := coerceToMap(m)
	assert.Equal(t, "value", result["key"])
}

func TestCoerceToMap_Struct(t *testing.T) {
	type Sample struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	result := coerceToMap(Sample{Name: "test", Age: 25})
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, float64(25), result["age"]) // JSON numbers are float64
}

// --- send ---

func TestSend_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("X-GH-Webhook-Signature"))

		var payload WebhookPayload
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "test.event", payload.EventType)
		assert.Equal(t, "user-1", payload.UserUUID)
		assert.Equal(t, "sandbox", payload.Environment)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	m := NewManager(server.URL, "test-secret", nil, nil, "")
	err := m.send("test.event", "user-1", map[string]interface{}{"data": "test"})
	assert.NoError(t, err)
}

func TestSend_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	m := NewManager(server.URL, "test-secret", nil, nil, "")
	err := m.send("test.event", "user-1", map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestSend_NoURL(t *testing.T) {
	m := NewManager("", "test-secret", nil, nil, "")
	err := m.send("test.event", "user-1", map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no webhook URL")
}

func TestSendAsync_NilQueue(t *testing.T) {
	m := NewManager("http://some.url", "secret", nil, nil, "")
	// No panic with nil queue
	m.SendAsync("test.event", "user-1", map[string]interface{}{}, 0)
}
