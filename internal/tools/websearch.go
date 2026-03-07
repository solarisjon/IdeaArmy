// Package tools provides external tool implementations for agent use.
package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/yourusername/ai-agent-team/internal/llm"
)

const (
	firecrawlSearchURL = "https://api.firecrawl.dev/v1/search"
	defaultMaxResults  = 5
)

// WebSearchTool returns the ToolDefinition for the web_search tool.
// Register this with BaseAgent.RegisterTool along with WebSearchExecutor.
func WebSearchTool() llm.ToolDefinition {
	params := json.RawMessage(`{
		"type": "object",
		"properties": {
			"query": {
				"type": "string",
				"description": "The search query to look up on the web"
			},
			"max_results": {
				"type": "integer",
				"description": "Maximum number of results to return (default 5)",
				"default": 5
			}
		},
		"required": ["query"]
	}`)

	return llm.ToolDefinition{
		Name:        "web_search",
		Description: "Search the web for current information, news, research, and real-world data. Use this to ground ideas in factual, up-to-date knowledge.",
		Parameters:  params,
	}
}

// SearchResult is a structured web search result surfaced to callers.
type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Snippet     string `json:"snippet"` // first 300 chars of markdown content
	Query       string `json:"query"`   // the search query that produced this result
}

// WebSearchExecutor returns a tool executor function for web_search.
// If FIRECRAWL_API_KEY is not set, it returns a graceful fallback message.
// notify is called with a human-readable status line before each search.
// onResults is called with structured results after each successful search;
// it may be nil if the caller only needs the formatted text.
func WebSearchExecutor(notify func(string), onResults func([]SearchResult)) func(arguments string) (string, error) {
	httpClient := &http.Client{Timeout: 30 * time.Second}

	return func(arguments string) (string, error) {
		var args struct {
			Query      string `json:"query"`
			MaxResults int    `json:"max_results"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return "", fmt.Errorf("invalid web_search arguments: %w", err)
		}
		if args.MaxResults <= 0 {
			args.MaxResults = defaultMaxResults
		}

		if notify != nil {
			notify(fmt.Sprintf("🔍 Searching web: %s", args.Query))
		}

		apiKey := os.Getenv("FIRECRAWL_API_KEY")
		if apiKey == "" {
			return fmt.Sprintf("[Web search unavailable: FIRECRAWL_API_KEY not set. Query was: %q — using internal knowledge instead.]", args.Query), nil
		}

		text, results, err := searchFirecrawl(httpClient, apiKey, args.Query, args.MaxResults)
		if err != nil {
			return text, err
		}
		if onResults != nil && len(results) > 0 {
			onResults(results)
		}
		return text, nil
	}
}

// firecrawlRequest is the Firecrawl search API request body.
type firecrawlRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

// firecrawlResult is a single Firecrawl search result.
type firecrawlResult struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Markdown    string `json:"markdown"`
}

// firecrawlResponse is the Firecrawl search API response.
type firecrawlResponse struct {
	Success bool              `json:"success"`
	Data    []firecrawlResult `json:"data"`
	Error   string            `json:"error"`
}

func searchFirecrawl(client *http.Client, apiKey, query string, maxResults int) (string, []SearchResult, error) {
	reqBody := firecrawlRequest{Query: query, Limit: maxResults}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("firecrawl: marshal error: %w", err)
	}

	req, err := http.NewRequest("POST", firecrawlSearchURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, fmt.Errorf("firecrawl: request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("[Web search failed: %v — using internal knowledge instead.]", err), nil, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("[Web search read error: %v — using internal knowledge.]", err), nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("[Web search API error (status %d) — using internal knowledge.]", resp.StatusCode), nil, nil
	}

	var result firecrawlResponse
	if err := json.Unmarshal(body, &result); err != nil || !result.Success {
		return fmt.Sprintf("[Web search parse error — using internal knowledge. Error: %s]", result.Error), nil, nil
	}

	if len(result.Data) == 0 {
		return fmt.Sprintf("[No web results found for: %q — using internal knowledge.]", query), nil, nil
	}

	structured := toSearchResults(query, result.Data)
	return formatResults(query, result.Data), structured, nil
}

// toSearchResults converts raw Firecrawl results to the public SearchResult type.
func toSearchResults(query string, data []firecrawlResult) []SearchResult {
	out := make([]SearchResult, 0, len(data))
	for _, r := range data {
		snippet := r.Markdown
		if len(snippet) > 300 {
			snippet = snippet[:300] + "…"
		}
		out = append(out, SearchResult{
			Title:       r.Title,
			URL:         r.URL,
			Description: r.Description,
			Snippet:     snippet,
			Query:       query,
		})
	}
	return out
}

func formatResults(query string, results []firecrawlResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Web search results for: %q\n\n", query))

	for i, r := range results {
		sb.WriteString(fmt.Sprintf("--- Result %d ---\n", i+1))
		if r.Title != "" {
			sb.WriteString(fmt.Sprintf("Title: %s\n", r.Title))
		}
		if r.URL != "" {
			sb.WriteString(fmt.Sprintf("URL: %s\n", r.URL))
		}
		if r.Description != "" {
			sb.WriteString(fmt.Sprintf("Summary: %s\n", r.Description))
		}
		if r.Markdown != "" {
			// Trim long content
			content := r.Markdown
			if len(content) > 800 {
				content = content[:800] + "..."
			}
			sb.WriteString(fmt.Sprintf("Content:\n%s\n", content))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
