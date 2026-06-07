package config

import (
	"os"
	"testing"
)

func TestParseConfig(t *testing.T) {
	paramsJSON := []byte(`{
		"runtime": {"name": "yunhu", "account_id": "main"},
		"config": {
			"token": "test_token_abcd",
			"webhook_port": 18080,
			"webhook_path": "/webhook/yunhu",
			"webhook_public_url": "https://example.com"
		}
	}`)

	sp, err := ParseStartParams(paramsJSON)
	if err != nil {
		t.Fatal(err)
	}
	if sp.Runtime.Name != "yunhu" {
		t.Errorf("expected runtime name yunhu, got %s", sp.Runtime.Name)
	}
	if sp.Runtime.AccountID != "main" {
		t.Errorf("expected account main, got %s", sp.Runtime.AccountID)
	}

	cfg, err := sp.ParseConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Token != "test_token_abcd" {
		t.Errorf("expected token test_token_abcd, got %s", cfg.Token)
	}
	if cfg.WebhookPort != 18080 {
		t.Errorf("expected port 18080, got %d", cfg.WebhookPort)
	}
	if cfg.WebhookPath != "/webhook/yunhu" {
		t.Errorf("expected path /webhook/yunhu, got %s", cfg.WebhookPath)
	}
}

func TestParseConfig_Defaults(t *testing.T) {
	paramsJSON := []byte(`{
		"runtime": {"name": "yunhu", "account_id": "main"},
		"config": {"token": "tok123"}
	}`)

	sp, err := ParseStartParams(paramsJSON)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := sp.ParseConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WebhookPort != 18080 {
		t.Errorf("expected default port 18080, got %d", cfg.WebhookPort)
	}
	if cfg.WebhookPath != "/webhook/yunhu" {
		t.Errorf("expected default path /webhook/yunhu, got %s", cfg.WebhookPath)
	}
}

func TestParseConfig_EnvFallback(t *testing.T) {
	os.Setenv("YUNHU_BOT_TOKEN", "env_token_xyz")
	defer os.Unsetenv("YUNHU_BOT_TOKEN")

	paramsJSON := []byte(`{
		"runtime": {"name": "yunhu", "account_id": "main"},
		"config": {}
	}`)

	sp, err := ParseStartParams(paramsJSON)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := sp.ParseConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Token != "env_token_xyz" {
		t.Errorf("expected env token, got %s", cfg.Token)
	}
}

func TestParseConfig_NoToken(t *testing.T) {
	os.Unsetenv("YUNHU_BOT_TOKEN")

	paramsJSON := []byte(`{
		"runtime": {"name": "yunhu", "account_id": "main"},
		"config": {}
	}`)

	sp, err := ParseStartParams(paramsJSON)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := sp.ParseConfig()
	if err != nil {
		t.Fatal(err)
	}

	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for missing token")
	}
}

func TestRedactedToken(t *testing.T) {
	cfg := &Config{Token: "abcdefgh"}
	redacted := cfg.RedactedToken()
	if redacted != "abcd****" {
		t.Errorf("expected abc****, got %s", redacted)
	}

	cfg2 := &Config{Token: "ab"}
	redacted2 := cfg2.RedactedToken()
	if redacted2 != "****" {
		t.Errorf("expected ****, got %s", redacted2)
	}
}

func TestListenAddr(t *testing.T) {
	cfg := &Config{WebhookPort: 8080}
	addr := cfg.ListenAddr()
	if addr != "0.0.0.0:8080" {
		t.Errorf("expected 0.0.0.0:8080, got %s", addr)
	}
}
