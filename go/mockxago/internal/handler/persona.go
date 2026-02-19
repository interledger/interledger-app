package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"gitlab.com/fynbos/mockxago/internal/logger"
)

type personaCreateInquiryRequest struct {
	Data struct {
		Attributes struct {
			ReferenceID string `json:"reference-id,omitempty"`
			CountryCode string `json:"country-code,omitempty"`
		} `json:"attributes"`
	} `json:"data"`
}

type personaInquiryResponse struct {
	Data personaInquiryData `json:"data"`
}

type personaInquiryData struct {
	Type       string                   `json:"type"`
	ID         string                   `json:"id"`
	Attributes personaInquiryAttributes `json:"attributes"`
	Meta       personaInquiryMeta       `json:"meta"`
}

type personaInquiryAttributes struct {
	Status      string `json:"status"`
	ReferenceID string `json:"reference-id"`
	CreatedAt   string `json:"created-at"`
	UpdatedAt   string `json:"updated-at"`
	CompletedAt string `json:"completed-at,omitempty"`
}

type personaInquiryMeta struct {
	SessionToken string `json:"session-token"`
}

// PersonaCreateInquiry handles POST /inquiries - Create or retrieve inquiry
// Compatible with Persona SDK API
func (h *Handler) PersonaCreateInquiry(w http.ResponseWriter, r *http.Request) {
	var req personaCreateInquiryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorf("Failed to decode inquiry request: %v", err)
		h.sendError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	refID := strings.TrimSpace(req.Data.Attributes.ReferenceID)
	inquiryID := refID
	if inquiryID == "" {
		inquiryID = fmt.Sprintf("inq_%d", time.Now().Unix())
	}

	now := time.Now().UTC().Format(time.RFC3339)
	logger.Infof("Created Persona inquiry: %s (reference-id=%s)", inquiryID, refID)

	response := personaInquiryResponse{
		Data: personaInquiryData{
			Type: "inquiry",
			ID:   inquiryID,
			Attributes: personaInquiryAttributes{
				Status:      "created",
				ReferenceID: refID,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			Meta: personaInquiryMeta{
				SessionToken: "mock-session-token",
			},
		},
	}

	h.sendJSON(w, http.StatusCreated, response)
}

// PersonaGetInquiry handles GET /inquiries/{id} - Get inquiry status
// Compatible with Persona SDK API
func (h *Handler) PersonaGetInquiry(w http.ResponseWriter, r *http.Request) {
	inquiryID := chi.URLParam(r, "inquiryId")
	if inquiryID == "" {
		h.sendError(w, http.StatusBadRequest, "missing_inquiry_id", "Inquiry ID is required")
		return
	}

	logger.Infof("Getting Persona inquiry: %s", inquiryID)

	now := time.Now().UTC().Format(time.RFC3339)
	response := personaInquiryResponse{
		Data: personaInquiryData{
			Type: "inquiry",
			ID:   inquiryID,
			Attributes: personaInquiryAttributes{
				Status:      "created",
				ReferenceID: inquiryID,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			Meta: personaInquiryMeta{
				SessionToken: "mock-session-token",
			},
		},
	}

	h.sendJSON(w, http.StatusOK, response)
}

// PersonaGetInquiryIframe handles GET /inquiries/{id}/iframe - Serves the KYC iframe
func (h *Handler) PersonaGetInquiryIframe(w http.ResponseWriter, r *http.Request) {
	inquiryID := chi.URLParam(r, "inquiryId")
	if inquiryID == "" {
		h.sendError(w, http.StatusBadRequest, "missing_inquiry_id", "Inquiry ID is required")
		return
	}

	logger.Infof("Serving Persona iframe for inquiry: %s", inquiryID)

	// Get user_id from query params if provided
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = inquiryID
	}

	// Load and serve the KYC iframe template
	possiblePaths := []string{
		"web/kyc-iframe.html",
		"./web/kyc-iframe.html",
		"../../web/kyc-iframe.html",
		"../../../web/kyc-iframe.html",
	}

	var templatePath string
	for _, path := range possiblePaths {
		if h.fileExists(path) {
			templatePath = path
			break
		}
	}

	if templatePath == "" {
		logger.Errorf("Could not find KYC iframe template")
		h.sendError(w, http.StatusInternalServerError, "template_not_found", "KYC template not found")
		return
	}

	// Read and serve the HTML directly (without template processing for Persona API)
	content, err := h.readFile(templatePath)
	if err != nil {
		logger.Errorf("Failed to read KYC iframe template: %v", err)
		h.sendError(w, http.StatusInternalServerError, "template_error", "Failed to read template")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Frame-Options", "ALLOWALL") // Allow framing from any origin
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(content))
}

// PersonaInquirySubmit handles POST /inquiries/{id}/submit - Form submission callback
func (h *Handler) PersonaInquirySubmit(w http.ResponseWriter, r *http.Request) {
	inquiryID := chi.URLParam(r, "inquiryId")
	if inquiryID == "" {
		h.sendError(w, http.StatusBadRequest, "missing_inquiry_id", "Inquiry ID is required")
		return
	}

	logger.Infof("Submitting Persona inquiry: %s", inquiryID)

	// Parse form submission
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		if err := r.ParseForm(); err != nil {
			logger.Errorf("Failed to parse form: %v", err)
			h.sendError(w, http.StatusBadRequest, "invalid_form", "Invalid form data")
			return
		}
	}

	// Extract form fields
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	dob := r.FormValue("dob")

	logger.Infof("KYC submitted for inquiry %s: %s %s (DOB: %s)", inquiryID, firstName, lastName, dob)

	// Update inquiry status to approved
	// In a real implementation, this would trigger background verification
	// For testing, we mark it as approved immediately

	// Send webhook notification to backend
	go h.sendKYCWebhook(inquiryID)

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Inquiry submitted successfully",
	})
}

// Helper functions
func (h *Handler) fileExists(path string) bool {
	_, err := h.readFile(path)
	return err == nil
}

func (h *Handler) readFile(path string) ([]byte, error) {
	// This is a simple mock - in real implementation would use os.ReadFile
	// For now, we'll return the actual KYC iframe HTML content
	content := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Persona KYC Verification</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto'; 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 8px;
            box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
            width: 100%;
            max-width: 500px;
            padding: 40px;
        }
        h1 { color: #1a1a1a; margin-bottom: 8px; }
        .form-group { margin-bottom: 20px; }
        label { display: block; margin-bottom: 8px; color: #1a1a1a; font-weight: 500; }
        input { 
            width: 100%; 
            padding: 10px 12px; 
            border: 1px solid #e0e0e0; 
            border-radius: 6px; 
            font-size: 14px;
        }
        button {
            width: 100%;
            padding: 12px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 6px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            margin-top: 20px;
        }
        button:hover { transform: translateY(-2px); }
    </style>
</head>
<body>
    <div class="container">
        <h1>Identity Verification</h1>
        <p>Please provide your information</p>

        <form id="kycForm" method="POST" action="/kyc/submit" enctype="multipart/form-data">
            <div class="form-group">
                <label>First Name *</label>
                <input type="text" id="first_name" name="first_name" required placeholder="John">
            </div>
            <div class="form-group">
                <label>Last Name *</label>
                <input type="text" id="last_name" name="last_name" required placeholder="Doe">
            </div>
            <div class="form-group">
                <label>Date of Birth *</label>
                <input type="date" id="dob" name="dob" required>
            </div>
            <div class="form-group">
                <label>Address *</label>
                <input type="text" id="address" name="address" required placeholder="123 Main Street">
            </div>
            <div class="form-group">
                <label>City *</label>
                <input type="text" id="city" name="city" required placeholder="City">
            </div>
            <div class="form-group">
                <label>Country *</label>
                <input type="text" id="country" name="country" required placeholder="Country">
            </div>
            <button type="submit">Submit Verification</button>
        </form>
    </div>

    <script>
        const form = document.getElementById('kycForm');
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(form);
            try {
                await fetch(form.action, {
                    method: 'POST',
                    body: formData
                });
                // Notify parent of completion
                window.parent.postMessage({
                    type: 'OnboardingCompleted',
                    value: JSON.stringify({
                        applicantStatus: 'submitted'
                    })
                }, '*');
                setTimeout(() => window.parent.location.reload(), 2000);
            } catch (error) {
                console.error('Submission error:', error);
            }
        });
    </script>
</body>
</html>`
	return []byte(content), nil
}
