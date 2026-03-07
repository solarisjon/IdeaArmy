package llmfactory

import (
	"fmt"

	"github.com/yourusername/ai-agent-team/internal/claude"
	"github.com/yourusername/ai-agent-team/internal/llm"
	"github.com/yourusername/ai-agent-team/internal/openai"
)

// NewClient creates the appropriate LLM client based on the resolved backend config.
func NewClient(cfg *llm.BackendConfig) (llm.Client, error) {
	switch cfg.Backend {
	case "anthropic":
		c := claude.NewClient(cfg.APIKey)
		c.Model = cfg.Model
		c.BaseURL = cfg.BaseURL
		return c, nil
	case "openai":
		c := openai.NewClient(cfg.APIKey, cfg.BaseURL, cfg.Model)
		c.User = cfg.User
		return c, nil
	default:
		return nil, fmt.Errorf("unknown backend: %q (use \"anthropic\" or \"openai\")", cfg.Backend)
	}
}

// NewClientAuto resolves the backend from env vars and creates the client.
// apiKeyOverride is optional — pass "" to rely entirely on env vars.
func NewClientAuto(apiKeyOverride string) (llm.Client, error) {
	cfg, err := llm.ResolveBackend(apiKeyOverride)
	if err != nil {
		return nil, err
	}
	return NewClient(cfg)
}

// NewClientWithModel creates a client identical to one from cfg but using the
// specified model. If model is empty, the default from cfg is used.
func NewClientWithModel(cfg *llm.BackendConfig, model string) (llm.Client, error) {
	override := *cfg // shallow copy
	if model != "" {
		override.Model = model
	}
	return NewClient(&override)
}

// ResolveBackendAuto is a convenience wrapper around llm.ResolveBackend.
func ResolveBackendAuto(apiKeyOverride string) (*llm.BackendConfig, error) {
	return llm.ResolveBackend(apiKeyOverride)
}
