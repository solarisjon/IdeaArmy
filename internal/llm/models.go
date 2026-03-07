package llm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

// ModelInfo describes an available LLM model.
type ModelInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	OwnedBy string `json:"owned_by"`
}

// ListModels queries the backend for available models.
// For OpenAI-compatible APIs it calls GET {BaseURL}/models.
// For Anthropic it returns a curated static list.
func ListModels(cfg *BackendConfig) ([]ModelInfo, error) {
	switch cfg.Backend {
	case "openai":
		return listOpenAIModels(cfg)
	case "anthropic":
		return listAnthropicModels(), nil
	default:
		return nil, fmt.Errorf("unsupported backend for model listing: %q", cfg.Backend)
	}
}

// listOpenAIModels calls the /models endpoint on an OpenAI-compatible API.
func listOpenAIModels(cfg *BackendConfig) ([]ModelInfo, error) {
	url := strings.TrimSuffix(cfg.BaseURL, "/") + "/models"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating models request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := NewHTTPClient(DefaultTimeout)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("models endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			ID      string `json:"id"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding models response: %w", err)
	}

	models := make([]ModelInfo, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, ModelInfo{
			ID:      m.ID,
			Name:    m.ID,
			OwnedBy: m.OwnedBy,
		})
	}

	sort.Slice(models, func(i, j int) bool { return models[i].ID < models[j].ID })
	return models, nil
}

// listAnthropicModels returns a curated list of known Anthropic models.
func listAnthropicModels() []ModelInfo {
	return []ModelInfo{
		{ID: "claude-opus-4-20250514", Name: "Claude Opus 4", OwnedBy: "anthropic"},
		{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", OwnedBy: "anthropic"},
		{ID: "claude-haiku-3-5-20241022", Name: "Claude Haiku 3.5", OwnedBy: "anthropic"},
		{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4 (default)", OwnedBy: "anthropic"},
	}
}
