package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockpti/internal/jobs"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/models"
)

func TestSDKScript_ReturnsJavaScript(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/sdk/index.js", nil)
	rr := httptest.NewRecorder()

	h.SDKScript(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "application/javascript") {
		t.Fatalf("expected javascript content-type, got %q", contentType)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "window.PTI") {
		t.Fatalf("expected PTI global in sdk script, got %q", body)
	}
	if !strings.Contains(body, "document.createElement('iframe')") {
		t.Fatalf("expected iframe creation in sdk script")
	}
	if !strings.Contains(body, "ptiFormsUrl") {
		t.Fatalf("expected forms URL usage in sdk script")
	}
}

func TestFormsLanding_ReturnsHTML(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/forms", nil)
	rr := httptest.NewRecorder()

	h.FormsLanding(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("expected html content-type, got %q", contentType)
	}
	if !strings.Contains(rr.Body.String(), "Mock PTI Embedded Form") {
		t.Fatalf("expected embedded forms heading")
	}
	if !strings.Contains(rr.Body.String(), "window.parent.postMessage") {
		t.Fatalf("expected forms landing page body")
	}
}

func TestCompleteAssessmentFromForm_SchedulesWebhook(t *testing.T) {
	h := newTestHandler()
	queue := jobs.NewQueue(h.store)
	h.SetQueue(queue)

	user := &models.User{ID: "user-1", CreatedAt: time.Now()}
	if err := h.store.SaveUser(context.Background(), user); err != nil {
		t.Fatalf("save user: %v", err)
	}

	body, _ := json.Marshal(map[string]string{
		"userId":      "user-1",
		"requestId":   "req-123",
		"dateOfBirth": "1990-01-15",
	})
	req := httptest.NewRequest(http.MethodPost, "/forms/complete", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.CompleteAssessmentFromForm(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	assessment, err := h.store.GetLatestAssessment(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("expected saved assessment, got error: %v", err)
	}
	if assessment.Assessment != "ACCEPTED" {
		t.Fatalf("expected ACCEPTED assessment, got %q", assessment.Assessment)
	}

	updatedUser, err := h.store.GetUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("expected updated user, got error: %v", err)
	}
	if updatedUser.DateOfBirth != "1990-01-15" {
		t.Fatalf("expected dateOfBirth to be updated, got %q", updatedUser.DateOfBirth)
	}

	readyJobs, err := queue.GetReadyJobs(10)
	if err != nil {
		t.Fatalf("get ready jobs: %v", err)
	}
	if len(readyJobs) != 1 {
		t.Fatalf("expected 1 webhook job, got %d", len(readyJobs))
	}
	if readyJobs[0].JobType != jobs.JobTypeUserAssessmentWebhook {
		t.Fatalf("expected %s job type, got %s", jobs.JobTypeUserAssessmentWebhook, readyJobs[0].JobType)
	}
}

func TestCompleteAssessmentFromForm_RequiresFields(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/forms/complete", strings.NewReader(`{"userId":""}`))
	rr := httptest.NewRecorder()

	h.CompleteAssessmentFromForm(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestCompleteAssessmentFromForm_RequiresDOBWhenMissingOnUser(t *testing.T) {
	h := newTestHandler()
	queue := jobs.NewQueue(h.store)
	h.SetQueue(queue)

	user := &models.User{ID: "user-2", CreatedAt: time.Now()}
	if err := h.store.SaveUser(context.Background(), user); err != nil {
		t.Fatalf("save user: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/forms/complete", strings.NewReader(`{"userId":"user-2","requestId":"req-456"}`))
	rr := httptest.NewRecorder()

	h.CompleteAssessmentFromForm(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}
