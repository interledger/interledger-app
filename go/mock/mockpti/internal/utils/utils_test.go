package utils

import "testing"

func TestGenerateUUID(t *testing.T) {
	id := GenerateUUID()
	if id == "" {
		t.Error("expected non-empty UUID")
	}
	if len(id) != 36 {
		t.Errorf("expected UUID length 36, got %d", len(id))
	}
}

func TestGenerateUUID_Unique(t *testing.T) {
	id1 := GenerateUUID()
	id2 := GenerateUUID()
	if id1 == id2 {
		t.Error("expected unique UUIDs")
	}
}

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if len(token) != 64 {
		t.Errorf("expected token length 64, got %d", len(token))
	}
}

func TestGenerateToken_Unique(t *testing.T) {
	t1, _ := GenerateToken()
	t2, _ := GenerateToken()
	if t1 == t2 {
		t.Error("expected unique tokens")
	}
}

func TestGenerateTokenExpiresAt(t *testing.T) {
	exp := GenerateTokenExpiresAt()
	if exp.IsZero() {
		t.Error("expected non-zero expiration time")
	}
}
