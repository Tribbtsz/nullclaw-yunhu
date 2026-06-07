package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

// Config holds the plugin configuration parsed from the start request.
type Config struct {
	Token            string `json:"token"`
	WebhookPort      int    `json:"webhook_port"`
	WebhookPath      string `json:"webhook_path"`
	WebhookPublicURL string `json:"webhook_public_url"`
}

// Runtime holds the runtime metadata passed by nullclaw in the start request.
type Runtime struct {
	Name      string `json:"name"`
	AccountID string `json:"account_id"`
	StateDir  string `json:"state_dir"`
}

type StartParams struct {
	Runtime Runtime        `json:"runtime"`
	Config  json.RawMessage `json:"config"`
}

func ParseStartParams(paramsJSON []byte) (*StartParams, error) {
	var sp StartParams
	if err := json.Unmarshal(paramsJSON, &sp); err != nil {
		return nil, fmt.Errorf("parse start params: %w", err)
	}
	return &sp, nil
}

func (sp *StartParams) ParseConfig() (*Config, error) {
	var cfg Config
	if err := json.Unmarshal(sp.Config, &cfg); err != nil {
		return nil, fmt.Errorf("parse plugin config: %w", err)
	}

	if cfg.Token == "" {
		cfg.Token = os.Getenv("YUNHU_BOT_TOKEN")
	}

	if cfg.WebhookPort <= 0 {
		cfg.WebhookPort = 18080
	}
	if cfg.WebhookPath == "" {
		cfg.WebhookPath = "/webhook/yunhu"
	}

	return &cfg, nil
}

func (cfg *Config) Validate() error {
	if cfg.Token == "" {
		return fmt.Errorf("token is required (set in plugin_config_json or YUNHU_BOT_TOKEN env)")
	}
	return nil
}

func (cfg *Config) RedactedToken() string {
	if len(cfg.Token) < 4 {
		return "****"
	}
	return cfg.Token[:4] + "****"
}

func (cfg *Config) ListenAddr() string {
	return fmt.Sprintf("0.0.0.0:%d", cfg.WebhookPort)
}

func (cfg *Config) LogInfo() {
	slog.Info("config loaded",
		"webhook_port", cfg.WebhookPort,
		"webhook_path", cfg.WebhookPath,
		"webhook_public_url", cfg.WebhookPublicURL,
	)
	if cfg.WebhookPublicURL != "" {
		fullURL := cfg.WebhookPublicURL + cfg.WebhookPath
		slog.Info("please configure your webhook URL in Yunhu console: " + fullURL)
	}
}
