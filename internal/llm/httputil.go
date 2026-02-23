package llm

import (
	"crypto/tls"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// DefaultTimeout is the default HTTP client timeout for LLM API calls.
// Override with LLM_TIMEOUT env var (in seconds).
const DefaultTimeout = 5 * time.Minute

// NewHTTPClient creates an *http.Client that respects LLM_SKIP_TLS_VERIFY
// and LLM_TIMEOUT environment variables.
//
// Set LLM_SKIP_TLS_VERIFY=true to skip certificate verification (useful for
// lab/dev environments behind corporate proxies with internal CAs).
//
// Set LLM_TIMEOUT=300 to override the timeout in seconds (default: 300).
// The timeout parameter is used as a fallback if LLM_TIMEOUT is not set.
func NewHTTPClient(timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	v := strings.ToLower(os.Getenv("LLM_SKIP_TLS_VERIFY"))
	if v == "true" || v == "1" || v == "yes" {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = true
	}

	// Allow env var override for timeout
	if envTimeout := os.Getenv("LLM_TIMEOUT"); envTimeout != "" {
		if secs, err := strconv.Atoi(envTimeout); err == nil && secs > 0 {
			timeout = time.Duration(secs) * time.Second
		}
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}
