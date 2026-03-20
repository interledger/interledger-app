package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestParseSecretSupportsPrefixFormats(t *testing.T) {
	k1, err := ParseSecret("myprefix_bXlzZWNyZXQ=")
	if err != nil || string(k1) != "mysecret" {
		t.Fatalf("ParseSecret simple prefix failed: key=%q err=%v", string(k1), err)
	}

	k2, err := ParseSecret("multi_part_prefix_bXlzZWNyZXQ=")
	if err != nil || string(k2) != "mysecret" {
		t.Fatalf("ParseSecret multi underscore failed: key=%q err=%v", string(k2), err)
	}
}

func TestComputeSignatureMatchesManualHMAC(t *testing.T) {
	key := []byte("local-test-webhook-secret")
	svixID := "msg_abc"
	timestamp := "1234567890"
	body := []byte(`{"eventType":"charge.interac.completed"}`)

	got := ComputeSignature(key, svixID, timestamp, body)
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(svixID + "." + timestamp + "." + string(body)))
	want := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if got != want {
		t.Fatalf("signature mismatch: got %q want %q", got, want)
	}
}

func TestSendAddsSvixHeadersAndSignatureValidity(t *testing.T) {
	secret := "local_bG9jYWwtdGVzdC13ZWJob29rLXNlY3JldA=="
	key := []byte("local-test-webhook-secret")

	var capturedHeaders http.Header
	var capturedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		defer r.Body.Close()
		b, _ := io.ReadAll(r.Body)
		capturedBody = b
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewSender(&http.Client{Timeout: 2 * time.Second})
	payload := map[string]any{"eventType": "charge.interac.completed", "status": "completed"}
	if err := sender.Send(t.Context(), server.URL, secret, payload); err != nil {
		t.Fatalf("Send() unexpected error: %v", err)
	}

	if !strings.HasPrefix(capturedHeaders.Get("svix-id"), "msg_") {
		t.Fatalf("svix-id mismatch: %q", capturedHeaders.Get("svix-id"))
	}
	if !strings.HasPrefix(capturedHeaders.Get("svix-signature"), "v1,") {
		t.Fatalf("svix-signature mismatch: %q", capturedHeaders.Get("svix-signature"))
	}

	timestamp := capturedHeaders.Get("svix-timestamp")
	sig := strings.TrimPrefix(capturedHeaders.Get("svix-signature"), "v1,")
	expected := ComputeSignature(key, capturedHeaders.Get("svix-id"), timestamp, capturedBody)
	if sig != expected {
		t.Fatalf("signature validation mismatch: got %q want %q", sig, expected)
	}

	wrongExpected := ComputeSignature([]byte("wrong-secret"), capturedHeaders.Get("svix-id"), timestamp, capturedBody)
	if sig == wrongExpected {
		t.Fatalf("signature should not validate with wrong secret")
	}

	var body map[string]any
	_ = json.Unmarshal(capturedBody, &body)
	if _, hasData := body["data"]; hasData {
		t.Fatalf("payload should be flat, got data wrapper")
	}
}
