package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Sender struct {
	httpClient *http.Client
	now        func() time.Time
}

func NewSender(client *http.Client) *Sender {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	return &Sender{
		httpClient: client,
		now:        time.Now,
	}
}

func ParseSecret(secret string) ([]byte, error) {
	idx := strings.LastIndex(secret, "_")
	if idx <= 0 || idx >= len(secret)-1 {
		return nil, fmt.Errorf("invalid webhook secret format")
	}

	decoded, err := base64.StdEncoding.DecodeString(secret[idx+1:])
	if err != nil {
		return nil, fmt.Errorf("decode webhook secret: %w", err)
	}

	return decoded, nil
}

func ComputeSignature(key []byte, svixID string, ts string, body []byte) string {
	signedPayload := svixID + "." + ts + "." + string(body)
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(signedPayload))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (s *Sender) Send(ctx context.Context, webhookURL string, secret string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	key, err := ParseSecret(secret)
	if err != nil {
		return err
	}

	svixID := "msg_" + uuid.NewString()
	timestamp := strconv.FormatInt(s.now().UTC().Unix(), 10)
	signature := "v1," + ComputeSignature(key, svixID, timestamp, body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("build webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("svix-id", svixID)
	req.Header.Set("svix-timestamp", timestamp)
	req.Header.Set("svix-signature", signature)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("webhook response status: %d", resp.StatusCode)
	}

	return nil
}
