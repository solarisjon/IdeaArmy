package llm

import (
	"fmt"
	"os"
	"strings"
)

// BackendConfig holds the resolved backend settings.
type BackendConfig struct {
	Backend string // "anthropic" or "openai"
	APIKey  string
	BaseURL string
	Model   string
	User    string // optional user identifier (required by some proxies like NetApp LLM proxy)
}

// ResolveBackend auto-detects the LLM backend from environment variables.
//
// Priority:
//  1. Explicit LLM_BACKEND env var ("anthropic" or "openai")
//  2. If ANTHROPIC_API_KEY or ANTHROPIC_KEY is set → anthropic
//  3. If LLMPROXY_KEY or OPENAI_API_KEY is set → openai
//
// Additional env vars:
//   - LLM_BASE_URL  — override the API base URL
//   - LLM_MODEL     — override the default model
//   - LLM_API_KEY   — explicit API key (highest priority for key)
func ResolveBackend(apiKeyOverride string) (*BackendConfig, error) {
	cfg := &BackendConfig{}

	// Determine backend
	cfg.Backend = strings.ToLower(os.Getenv("LLM_BACKEND"))

	if cfg.Backend == "" {
		// Auto-detect from available keys
		switch {
		case apiKeyOverride != "":
			// If caller passed a key, check if it looks like an Anthropic key
			if strings.HasPrefix(apiKeyOverride, "sk-ant-") {
				cfg.Backend = "anthropic"
			} else {
				cfg.Backend = "anthropic" // default assumption for passed keys
			}
		case os.Getenv("ANTHROPIC_API_KEY") != "" || os.Getenv("ANTHROPIC_KEY") != "":
			cfg.Backend = "anthropic"
		case os.Getenv("LLMPROXY_KEY") != "" || os.Getenv("OPENAI_API_KEY") != "" || os.Getenv("LLM_API_KEY") != "":
			cfg.Backend = "openai"
		default:
			return nil, fmt.Errorf("no LLM API key found. Set ANTHROPIC_API_KEY, LLMPROXY_KEY, OPENAI_API_KEY, or LLM_API_KEY")
		}
	}

	// Resolve API key (priority: LLM_API_KEY > apiKeyOverride > backend-specific)
	cfg.APIKey = os.Getenv("LLM_API_KEY")
	if cfg.APIKey == "" && apiKeyOverride != "" {
		cfg.APIKey = apiKeyOverride
	}
	if cfg.APIKey == "" {
		switch cfg.Backend {
		case "anthropic":
			cfg.APIKey = os.Getenv("ANTHROPIC_API_KEY")
			if cfg.APIKey == "" {
				cfg.APIKey = os.Getenv("ANTHROPIC_KEY")
			}
		case "openai":
			cfg.APIKey = os.Getenv("OPENAI_API_KEY")
			if cfg.APIKey == "" {
				proxyKey := os.Getenv("LLMPROXY_KEY")
				cfg.APIKey = parseLLMProxyKey(proxyKey)
				cfg.User = parseLLMProxyUser(proxyKey)
			}
		}
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("no API key found for backend %q", cfg.Backend)
	}

	// Resolve base URL
	cfg.BaseURL = os.Getenv("LLM_BASE_URL")
	if cfg.BaseURL == "" {
		switch cfg.Backend {
		case "anthropic":
			cfg.BaseURL = "https://api.anthropic.com/v1/messages"
		case "openai":
			cfg.BaseURL = "https://llm-proxy-api.ai.eng.netapp.com/v1"
		}
	}

	// Resolve model
	cfg.Model = os.Getenv("LLM_MODEL")
	if cfg.Model == "" {
		switch cfg.Backend {
		case "anthropic":
			cfg.Model = "claude-sonnet-4-20250514"
		case "openai":
			cfg.Model = "gpt-4o"
		}
	}

	return cfg, nil
}

// parseLLMProxyKey extracts the key from "user=xxx&key=sk_xxx" format.
func parseLLMProxyKey(raw string) string {
	if raw == "" {
		return ""
	}
	for _, part := range strings.Split(raw, "&") {
		if strings.HasPrefix(part, "key=") {
			return strings.TrimPrefix(part, "key=")
		}
	}
	// If no key= prefix, return the whole value
	return raw
}

// parseLLMProxyUser extracts the user from "user=xxx&key=sk_xxx" format.
func parseLLMProxyUser(raw string) string {
	if raw == "" {
		return ""
	}
	for _, part := range strings.Split(raw, "&") {
		if strings.HasPrefix(part, "user=") {
			return strings.TrimPrefix(part, "user=")
		}
	}
	return ""
}
