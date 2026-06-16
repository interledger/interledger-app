package client

import (
	"testing"

	"gitlab.com/fynbos/backend/providers/plaid"
)

func TestBuildSDKConfig_APIURLOverride(t *testing.T) {
	cfg, err := buildSDKConfig(plaid.Config{
		Env:    "sandbox",
		APIURL: "http://mockplaid:8080",
	})
	if err != nil {
		t.Fatalf("buildSDKConfig() error: %v", err)
	}
	if len(cfg.Servers) != 1 || cfg.Servers[0].URL != "http://mockplaid:8080" {
		t.Fatalf("Servers not overridden: %+v", cfg.Servers)
	}
}

func TestBuildSDKConfig_SandboxDefault(t *testing.T) {
	cfg, err := buildSDKConfig(plaid.Config{Env: "sandbox"})
	if err != nil {
		t.Fatalf("buildSDKConfig() error: %v", err)
	}
	if len(cfg.Servers) != 1 || cfg.Servers[0].URL != "https://sandbox.plaid.com" {
		t.Fatalf("expected sandbox URL, got: %+v", cfg.Servers)
	}
}

func TestBuildSDKConfig_UnknownEnv(t *testing.T) {
	if _, err := buildSDKConfig(plaid.Config{Env: "bogus"}); err == nil {
		t.Fatal("expected error for unknown env, got nil")
	}
}
