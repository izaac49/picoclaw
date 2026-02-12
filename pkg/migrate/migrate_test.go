package migrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sipeed/picoclaw/pkg/config"
)

func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "apiKey", "api_key"},
		{"two words", "apiBase", "api_base"},
		{"three words", "maxToolIterations", "max_tool_iterations"},
		{"already snake", "api_key", "api_key"},
		{"single word", "enabled", "enabled"},
		{"all lower", "model", "model"},
		{"consecutive caps", "apiURL", "api_url"},
		{"starts upper", "Model", "model"},
		{"bridge url", "bridgeUrl", "bridge_url"},
		{"client id", "clientId", "client_id"},
		{"app secret", "appSecret", "app_secret"},
		{"verification token", "verificationToken", "verification_token"},
		{"allow from", "allowFrom", "allow_from"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := camelToSnake(tt.input)
			if got != tt.want {
				t.Errorf("camelToSnake(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConvertKeysToSnake(t *testing.T) {
	input := map[string]interface{}{
		"apiKey":  "test-key",
		"apiBase": "https://example.com",
		"nested": map[string]interface{}{
			"maxTokens":   float64(8192),
			"allowFrom":   []interface{}{"user1", "user2"},
			"deeperLevel": map[string]interface{}{
				"clientId": "abc",
			},
		},
	}

	result := convertKeysToSnake(input)
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map[string]interface{}")
	}

	if _, ok := m["api_key"]; !ok {
		t.Error("expected key 'api_key' after conversion")
	}
	if _, ok := m["api_base"]; !ok {
		t.Error("expected key 'api_base' after conversion")
	}

	nested, ok := m["nested"].(map[string]interface{})
	if !ok {
		t.Fatal("expected nested map")
	}
	if _, ok := nested["max_tokens"]; !ok {
		t.Error("expected key 'max_tokens' in nested map")
	}
	if _, ok := nested["allow_from"]; !ok {
		t.Error("expected key 'allow_from' in nested map")
	}

	deeper, ok := nested["deeper_level"].(map[string]interface{})
	if !ok {
		t.Fatal("expected deeper_level map")
	}
	if _, ok := deeper["client_id"]; !ok {
		t.Error("expected key 'client_id' in deeper level")
	}
}

func TestLoadOpenClawConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "openclaw.json")

	openclawConfig := map[string]interface{}{
		"providers": map[string]interface{}{
			"openrouter": map[string]interface{}{
				"apiKey":  "sk-or-test123",
				"apiBase": "https://openrouter.ai/api/v1",
			},
		},
		"agents": map[string]interface{}{
			"defaults": map[string]interface{}{
				"maxTokens": float64(4096),
				"model":     "meta-llama/llama-3-8b-instruct",
			},
		},
	}

	data, err := json.Marshal(openclawConfig)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	result, err := LoadOpenClawConfig(configPath)
	if err != nil {
		t.Fatalf("LoadOpenClawConfig: %v", err)
	}

	providers, ok := result["providers"].(map[string]interface{})
	if !ok {
		t.Fatal("expected providers map")
	}
	openrouter, ok := providers["openrouter"].(map[string]interface{})
	if !ok {
		t.Fatal("expected openrouter map")
	}
	if openrouter["api_key"] != "sk-or-test123" {
		t.Errorf("api_key = %v, want sk-or-test123", openrouter["api_key"])
	}

	agents, ok := result["agents"].(map[string]interface{})
	if !ok {
		t.Fatal("expected agents map")
	}
	defaults, ok := agents["defaults"].(map[string]interface{})
	if !ok {
		t.Fatal("expected defaults map")
	}
	if defaults["max_tokens"] != float64(4096) {
		t.Errorf("max_tokens = %v, want 4096", defaults["max_tokens"])
	}
}

func TestConvertConfig(t *testing.T) {
	t.Run("providers mapping", func(t *testing.T) {
		data := map[string]interface{}{
			"providers": map[string]interface{}{
				"openrouter": map[string]interface{}{
					"api_key": "sk-or-test",
				},
			},
		}

		cfg, warnings, err := ConvertConfig(data)
		if err != nil {
			t.Fatalf("ConvertConfig: %v", err)
		}
		if len(warnings) != 0 {
			t.Errorf("expected no warnings, got %v", warnings)
		}
		if cfg.Providers.OpenRouter.APIKey != "sk-or-test" {
			t.Errorf("OpenRouter.APIKey = %q, want %q", cfg.Providers.OpenRouter.APIKey, "sk-or-test")
		}
	})

	t.Run("unsupported provider warning", func(t *testing.T) {
		data := map[string]interface{}{
			"providers": map[string]interface{}{
				"deepseek": map[string]interface{}{
					"api_key": "sk-deep-test",
				},
			},
		}

		_, warnings, err := ConvertConfig(data)
		if err != nil {
			t.Fatalf("ConvertConfig: %v", err)
		}
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d", len(warnings))
		}
		if warnings[0] != "Provider 'deepseek' not supported in PicoClaw, skipping" {
			t.Errorf("unexpected warning: %s", warnings[0])
		}
	})

	t.Run("channels mapping", func(t *testing.T) {
		data := map[string]interface{}{
			"channels": map[string]interface{}{
				"telegram": map[string]interface{}{
					"enabled":    true,
					"token":      "tg-token-123",
					"allow_from": []interface{}{"user1"},
				},
				"discord": map[string]interface{}{
					"enabled": true,
					"token":   "disc-token-456",
				},
			},
		}

		cfg, _, err := ConvertConfig(data)
		if err != nil {
			t.Fatalf("ConvertConfig: %v", err)
		}
		if !cfg.Channels.Telegram.Enabled {
			t.Error("Telegram should be enabled")
		}
		if cfg.Channels.Telegram.Token != "tg-token-123" {
			t.Errorf("Telegram.Token = %q, want %q", cfg.Channels.Telegram.Token, "tg-token-123")
		}
		if len(cfg.Channels.Telegram.AllowFrom) != 1 || cfg.Channels.Telegram.AllowFrom[0] != "user1" {
			t.Errorf("Telegram.AllowFrom = %v, want [user1]", cfg.Channels.Telegram.AllowFrom)
		}
		if !cfg.Channels.Discord.Enabled {
			t.Error("Discord should be enabled")
		}
	})

	t.Run("unsupported channel warning", func(t *testing.T) {
		data := map[string]interface{}{
			"channels": map[string]interface{}{
				"email": map[string]interface{}{
					"enabled": true,
				},
			},
		}

		_, warnings, err := ConvertConfig(data)
		if err != nil {
			t.Fatalf("ConvertConfig: %v", err)
		}
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d", len(warnings))
		}
		if warnings[0] != "Channel 'email' not supported in PicoClaw, skipping" {
			t.Errorf("unexpected warning: %s", warnings[0])
		}
	})

	t.Run("agent defaults", func(t *testing.T) {
		data := map[string]interface{}{
			"agents": map[string]interface{}{
				"defaults": map[string]interface{}{
					"model":                "meta-llama/llama-3-8b-instruct",
					"max_tokens":           float64(4096),
					"temperature":          0.5,
					"max_tool_iterations":  float64(10),
					"workspace":            "~/.openclaw/workspace",
				},
			},
		}

		cfg, _, err := ConvertConfig(data)
		if err != nil {
			t.Fatalf("ConvertConfig: %v", err)
		}
		if cfg.Agents.Defaults.Model != "meta-llama/llama-3-8b-instruct" {
			t.Errorf("Model = %q, want %q", cfg.Agents.Defaults.Model, "meta-llama/llama-3-8b-instruct")
		}
		if cfg.Agents.Defaults.MaxTokens != 4096 {
			t.Errorf("MaxTokens = %d, want %d", cfg.Agents.Defaults.MaxTokens, 4096)
		}
		if cfg.Agents.Defaults.Temperature != 0.5 {
			t.Errorf("Temperature = %f, want %f", cfg.Agents.Defaults.Temperature, 0.5)
		}
		if cfg.Agents.Defaults.Workspace != "~/.picoclaw/workspace" {
			t.Errorf("Workspace = %q, want %q", cfg.Agents.Defaults.Workspace, "~/.picoclaw/workspace")
		}
	})

	t.Run("empty config", func(t *testing.T) {
		data := map[string]interface{}{}

		cfg, warnings, err := ConvertConfig(data)
		if err != nil {
			t.Fatalf("ConvertConfig: %v", err)
		}
		if len(warnings) != 0 {
			t.Errorf("expected no warnings, got %v", warnings)
		}
		if cfg.Agents.Defaults.Model != "openrouter" {
			t.Errorf("default model should be openrouter, got %q", cfg.Agents.Defaults.Model)
		}
	})
}

func TestMergeConfig(t *testing.T) {
	t.Run("fills empty fields", func(t *testing.T) {
		existing := config.DefaultConfig()
		incoming := config.DefaultConfig()
		incoming.Providers.OpenRouter.APIKey = "sk-or-incoming"

		result := MergeConfig(existing, incoming)
		if result.Providers.OpenRouter.APIKey != "sk-or-incoming" {
			t.Errorf("OpenRouter.APIKey = %q, want %q", result.Providers.OpenRouter.APIKey, "sk-or-incoming")
		}
	})

	t.Run("preserves existing non-empty fields", func(t *testing.T) {
		existing := config.DefaultConfig()
		existing.Providers.OpenRouter.APIKey = "sk-or-existing"

		incoming := config.DefaultConfig()
		incoming.Providers.OpenRouter.APIKey = "sk-or-incoming"

		result := MergeConfig(existing, incoming)
		if result.Providers.OpenRouter.APIKey != "sk-or-existing" {
			t.Errorf("OpenRouter.APIKey should be preserved, got %q", result.Providers.OpenRouter.APIKey)
		}
	})

	t.Run("merges enabled channels", func(t *testing.T) {
		existing := config.DefaultConfig()
		incoming := config.DefaultConfig()
		incoming.Channels.Telegram.Enabled = true
		incoming.Channels.Telegram.Token = "tg-token"

		result := MergeConfig(existing, incoming)
		if !result.Channels.Telegram.Enabled {
			t.Error("Telegram should be enabled after merge")
		}
		if result.Channels.Telegram.Token != "tg-token" {
			t.Errorf("Telegram.Token = %q, want %q", result.Channels.Telegram.Token, "tg-token")
		}
	})
}

func TestPlanWorkspaceMigration(t *testing.T) {
	t.Run("copies available files", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := t.TempDir()

		os.WriteFile(filepath.Join(srcDir, "AGENTS.md"), []byte("# Agents"), 0644)
		os.WriteFile(filepath.Join(srcDir, "SOUL.md"), []byte("# Soul"), 0644)
		os.WriteFile(filepath.Join(srcDir, "USER.md"), []byte("# User"), 0644)

		actions, err := PlanWorkspaceMigration(srcDir, dstDir, false)
		if err != nil {
			t.Fatalf("PlanWorkspaceMigration: %v", err)
		}

		copyCount := 0
		skipCount := 0
		for _, a := range actions {
			if a.Type == ActionCopy {
				copyCount++
			}
			if a.Type == ActionSkip {
				skipCount++
			}
		}
		if copyCount != 3 {
			t.Errorf("expected 3 copies, got %d", copyCount)
		}
	})
}

func TestFindOpenClawConfig(t *testing.T) {
	t.Run("finds openclaw.json", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "openclaw.json")
		os.WriteFile(configPath, []byte("{}"), 0644)

		found, err := findOpenClawConfig(tmpDir)
		if err != nil {
			t.Fatalf("findOpenClawConfig: %v", err)
		}
		if found != configPath {
			t.Errorf("found %q, want %q", found, configPath)
		}
	})
}

func TestRewriteWorkspacePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"default path", "~/.openclaw/workspace", "~/.picoclaw/workspace"},
		{"custom path", "/custom/path", "/custom/path"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rewriteWorkspacePath(tt.input)
			if got != tt.want {
				t.Errorf("rewriteWorkspacePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
