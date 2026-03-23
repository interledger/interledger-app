package external

import "testing"

func TestNew_UsesCHIMONEYAPIBaseURLOverride(t *testing.T) {
	t.Setenv("CHIMONEY_API_BASE_URL", "http://mockchimoney:8080/v0.2.4/")

	c, ok := New(nil).(*client)
	if !ok {
		t.Fatalf("New() returned unexpected client type")
	}

	if c.baseURL != "http://mockchimoney:8080/v0.2.4" {
		t.Fatalf("baseURL mismatch: got %q want %q", c.baseURL, "http://mockchimoney:8080/v0.2.4")
	}
}
