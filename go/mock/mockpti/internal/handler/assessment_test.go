package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/fynbos/mock/mockpti/internal/models"
)

func TestStartUserAssessment_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	// Create user first
	createBody := models.CreateUserRequest{ID: "user-assess-1", Type: "PERSON"}
	createPayload, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(createPayload))
	ptiHeaders(createReq)
	router.ServeHTTP(httptest.NewRecorder(), createReq)

	// Start assessment
	assessBody := models.StartAssessmentRequest{ID: "user-assess-1", Type: "PERSON"}
	assessPayload, _ := json.Marshal(assessBody)
	req := httptest.NewRequest(http.MethodPost, "/users/assessments", bytes.NewReader(assessPayload))
	ptiHeaders(req)
	req.Header.Set("x-pti-scenario-id", "ilf_withdrawal")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.ID == "" {
		t.Error("expected non-empty assessment request ID")
	}
}

func TestStartUserAssessment_UsesProvidedRequestID(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	// Create user
	createBody := models.CreateUserRequest{ID: "user-assess-2", Type: "PERSON"}
	createPayload, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(createPayload))
	ptiHeaders(createReq)
	router.ServeHTTP(httptest.NewRecorder(), createReq)

	// Start assessment with explicit request ID
	assessBody := models.StartAssessmentRequest{ID: "user-assess-2"}
	assessPayload, _ := json.Marshal(assessBody)
	req := httptest.NewRequest(http.MethodPost, "/users/assessments", bytes.NewReader(assessPayload))
	ptiHeaders(req)
	req.Header.Set("x-pti-request-id", "custom-req-id")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.ID != "custom-req-id" {
		t.Errorf("expected request ID custom-req-id, got %s", resp.ID)
	}
}

func TestStartUserAssessment_MissingUserID(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	body := models.StartAssessmentRequest{}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/users/assessments", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestStartUserAssessment_UserNotFound(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	body := models.StartAssessmentRequest{ID: "nonexistent"}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/users/assessments", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestGetUserAssessment_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	// Create user
	createBody := models.CreateUserRequest{ID: "user-getassess-1", Type: "PERSON"}
	createPayload, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(createPayload))
	ptiHeaders(createReq)
	router.ServeHTTP(httptest.NewRecorder(), createReq)

	// Start assessment
	assessBody := models.StartAssessmentRequest{ID: "user-getassess-1"}
	assessPayload, _ := json.Marshal(assessBody)
	assessReq := httptest.NewRequest(http.MethodPost, "/users/assessments", bytes.NewReader(assessPayload))
	ptiHeaders(assessReq)
	router.ServeHTTP(httptest.NewRecorder(), assessReq)

	// Get assessment
	req := httptest.NewRequest(http.MethodGet, "/users/user-getassess-1/assessments", nil)
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var assessment models.Assessment
	_ = json.NewDecoder(rr.Body).Decode(&assessment)

	if assessment.Assessment != "approved" {
		t.Errorf("expected assessment approved, got %s", assessment.Assessment)
	}
	if assessment.UserID != "user-getassess-1" {
		t.Errorf("expected user ID user-getassess-1, got %s", assessment.UserID)
	}
	if assessment.ResourceType != "assessment" {
		t.Errorf("expected resourceType assessment, got %s", assessment.ResourceType)
	}
	if assessment.Tier != 1 {
		t.Errorf("expected tier 1, got %d", assessment.Tier)
	}
}

func TestGetUserAssessment_NotFound(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/users/nonexistent/assessments", nil)
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}
