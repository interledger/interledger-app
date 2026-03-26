package external

import "testing"

func TestNew_TrimsTrailingSlashFromBaseURL(t *testing.T) {
	c, ok := New("http://mockchimoney:8080/v0.2.4/", "test-key", nil).(*client)
	if !ok {
		t.Fatalf("New() returned unexpected client type")
	}

	if c.baseURL != "http://mockchimoney:8080/v0.2.4" {
		t.Fatalf("baseURL mismatch: got %q want %q", c.baseURL, "http://mockchimoney:8080/v0.2.4")
	}
}
