// Package auth mints Ory Kratos session tokens for performance-test wallets.
//
// The backend's user interceptor accepts an "authorization: Bearer <token>"
// header and resolves it via Kratos ToSession (see backend/user/middleware).
// Kratos issues those tokens from its *native* (API) login flow, which is plain
// JSON — no browser and no CSRF handling required. That is what makes driving a
// hundred distinct wallets from a single process practical.
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	kratos "github.com/ory/kratos-client-go"
)

// ErrNoSessionToken is returned when Kratos accepts the credentials but does not
// hand back a token. That happens when the flow is treated as a browser flow, so
// it points at a misconfigured Kratos URL rather than at bad credentials.
var ErrNoSessionToken = errors.New("kratos returned no session token; check that target.kratos_url points at the public Kratos endpoint")

// Client logs wallets in against Kratos.
type Client struct {
	api *kratos.APIClient
}

// New builds a Client for the public Kratos endpoint at kratosURL.
func New(kratosURL string, timeout time.Duration) *Client {
	cfg := kratos.NewConfiguration()
	cfg.HTTPClient = &http.Client{Timeout: timeout}
	cfg.Servers = kratos.ServerConfigurations{
		{URL: kratosURL, Description: "Public Kratos"},
	}

	return &Client{api: kratos.NewAPIClient(cfg)}
}

// Login exchanges an email and password for a Kratos session token.
//
// A token, not a cookie: the backend accepts both, but tokens avoid per-wallet
// cookie jars and survive being passed around between goroutines.
func (c *Client) Login(ctx context.Context, email, password string) (string, error) {
	flow, resp, err := c.api.FrontendAPI.CreateNativeLoginFlow(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("create native login flow: %w", kratosError(resp, err))
	}

	body := kratos.UpdateLoginFlowWithPasswordMethodAsUpdateLoginFlowBody(
		kratos.NewUpdateLoginFlowWithPasswordMethod(email, "password", password),
	)

	login, resp, err := c.api.FrontendAPI.UpdateLoginFlow(ctx).
		Flow(flow.Id).
		UpdateLoginFlowBody(body).
		Execute()
	if err != nil {
		return "", fmt.Errorf("submit login for %s: %w", email, kratosError(resp, err))
	}

	token := login.GetSessionToken()
	if token == "" {
		return "", ErrNoSessionToken
	}

	return token, nil
}

// Verify confirms a token still resolves to a session, so a misconfigured wallet
// fails during setup rather than as a wall of Unauthenticated RPCs mid-run.
//
// This is the same ToSession call the backend interceptor makes, which means it
// also catches the AAL trap: a wallet with TOTP enrolled logs in at aal1, but
// Kratos is configured with session.whoami.required_aal: highest_available, so
// every RPC for that wallet would fail with an AAL2 error. Perf wallets must not
// have TOTP enrolled.
func (c *Client) Verify(ctx context.Context, token string) error {
	session, resp, err := c.api.FrontendAPI.ToSession(ctx).XSessionToken(token).Execute()
	if err != nil {
		return fmt.Errorf("resolve session: %w", kratosError(resp, err))
	}
	if !session.GetActive() {
		return errors.New("session is not active")
	}
	return nil
}

// kratosFlow is the subset of a Kratos flow response that carries the reason a
// submission was rejected.
type kratosFlow struct {
	UI struct {
		Messages []kratosMessage `json:"messages"`
		Nodes    []struct {
			Attributes struct {
				Name string `json:"name"`
			} `json:"attributes"`
			Messages []kratosMessage `json:"messages"`
		} `json:"nodes"`
	} `json:"ui"`
	Error *struct {
		Message string `json:"message"`
		Reason  string `json:"reason"`
	} `json:"error"`
}

type kratosMessage struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

// kratosError turns a Kratos SDK error into something readable.
//
// On a rejected submission Kratos replies 400 with the entire flow — several
// kilobytes of form definition wrapped around one sentence explaining what went
// wrong. Dumping that raw makes a failed run unreadable, so pull out the
// messages and leave the rest.
func kratosError(resp *http.Response, err error) error {
	body := errorBody(resp, err)
	if len(body) == 0 {
		return err
	}

	if reasons := flowMessages(body); len(reasons) > 0 {
		return fmt.Errorf("%w: %s", err, strings.Join(reasons, "; "))
	}

	// Not a flow response: include a bounded slice of the raw body rather than
	// discarding the only detail available.
	if len(body) > 512 {
		body = body[:512]
	}
	return fmt.Errorf("%w: %s", err, string(body))
}

func errorBody(resp *http.Response, err error) []byte {
	// The SDK captures the body on its own error type, which is more reliable than
	// re-reading the response.
	var apiErr *kratos.GenericOpenAPIError
	if errors.As(err, &apiErr) && len(apiErr.Body()) > 0 {
		return apiErr.Body()
	}

	if resp == nil || resp.Body == nil {
		return nil
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if readErr != nil {
		return nil
	}
	return body
}

// flowMessages extracts the human-readable errors from a Kratos flow body.
func flowMessages(body []byte) []string {
	var flow kratosFlow
	if err := json.Unmarshal(body, &flow); err != nil {
		return nil
	}

	var reasons []string

	if flow.Error != nil {
		if flow.Error.Message != "" {
			reasons = append(reasons, flow.Error.Message)
		}
		if flow.Error.Reason != "" && flow.Error.Reason != flow.Error.Message {
			reasons = append(reasons, flow.Error.Reason)
		}
	}

	for _, m := range flow.UI.Messages {
		if m.Type == "error" && m.Text != "" {
			reasons = append(reasons, m.Text)
		}
	}

	// Field-level errors say which trait Kratos objected to, which matters when
	// provisioning rejects a generated phone number or email.
	for _, node := range flow.UI.Nodes {
		for _, m := range node.Messages {
			if m.Type != "error" || m.Text == "" {
				continue
			}
			if name := node.Attributes.Name; name != "" {
				reasons = append(reasons, name+": "+m.Text)
				continue
			}
			reasons = append(reasons, m.Text)
		}
	}

	return reasons
}
