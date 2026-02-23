package llm

import (
	"crypto/tls"
	"net/http"
	"os"
	"strings"
	"time"
)

// NewHTTPClient creates an *http.Client that respects LLM_SKIP_TLS_VERIFY.
// Set LLM_SKIP_TLS_VERIFY=true to skip certificate verification (useful for
// lab/dev environments behind corporate proxies with internal CAs).
func NewHTTPClient(timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	v := strings.ToLower(os.Getenv("LLM_SKIP_TLS_VERIFY"))
	if v == "true" || v == "1" || v == "yes" {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = true
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}
