package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fork-svg/magpie-zero2xlab/config"
	"github.com/fork-svg/magpie-zero2xlab/internal/db"
	"github.com/fork-svg/magpie-zero2xlab/internal/models"
)

// Proxy handles proxying requests to zero2xlab
type Proxy struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewProxy creates a new proxy instance
func NewProxy(cfg *config.Config) *Proxy {
	return &Proxy{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for streaming
		},
	}
}

// ChatRequest represents an OpenAI-compatible chat request
type ChatRequest struct {
	Model            string        `json:"model"`
	Messages         []Message     `json:"messages"`
	Stream           bool          `json:"stream,omitempty"`
	Temperature      *float64      `json:"temperature,omitempty"`
	MaxTokens        *int          `json:"max_tokens,omitempty"`
	TopP             *float64      `json:"top_p,omitempty"`
	FrequencyPenalty *float64      `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64      `json:"presence_penalty,omitempty"`
	Stop             []string      `json:"stop,omitempty"`
	Tools            []interface{} `json:"tools,omitempty"`
	ToolChoice       interface{}   `json:"tool_choice,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// ForwardRequest forwards a chat completion request to zero2xlab
func (p *Proxy) ForwardRequest(w http.ResponseWriter, r *http.Request, apiKeyID uint) error {
	startTime := time.Now()

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	// Parse to get model name
	var chatReq ChatRequest
	if err := json.Unmarshal(body, &chatReq); err != nil {
		return fmt.Errorf("failed to parse request: %w", err)
	}

	// Build the target URL
	targetURL := fmt.Sprintf("%s/chat/completions", p.cfg.Zero2XLabBaseURL)

	// Create the proxied request
	proxyReq, err := http.NewRequest(r.Method, targetURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create proxy request: %w", err)
	}

	// Set headers
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.cfg.Zero2XLabAPIKey))

	// Copy other headers that should be forwarded
	for _, h := range []string{"Accept", "Accept-Encoding"} {
		if v := r.Header.Get(h); v != "" {
			proxyReq.Header.Set(h, v)
		}
	}

	// Send the request
	resp, err := p.httpClient.Do(proxyReq)
	if err != nil {
		return fmt.Errorf("failed to forward request: %w", err)
	}
	defer resp.Body.Close()

	// Log the request
	duration := time.Since(startTime).Milliseconds()

	// For streaming responses, we need to count tokens differently
	var tokens int
	if chatReq.Stream {
		// Stream the response
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Transfer-Encoding", "chunked")

		// Create a counting reader
		counter := &countingReader{reader: resp.Body}

		// Stream the response
		buf := make([]byte, 4096)
		for {
			n, err := counter.Read(buf)
			if n > 0 {
				w.Write(buf[:n])
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
		}

		// Estimate tokens from bytes (rough approximation)
		tokens = counter.count / 4
	} else {
		// Read the full response
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		// Parse response to count tokens
		var chatResp ChatResponse
		if err := json.Unmarshal(respBody, &chatResp); err == nil {
			tokens = chatResp.Usage.TotalTokens
		}

		// Copy response headers
		for k, v := range resp.Header {
			if !strings.HasPrefix(strings.ToLower(k), "connection") {
				w.Header()[k] = v
			}
		}

		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
	}

	// Update usage stats
	go func() {
		db.UpdateAPIKeyUsage(apiKeyID, tokens)
		db.CreateRequestLog(&models.RequestLog{
			APIKeyID:  apiKeyID,
			Method:    r.Method,
			Path:      r.URL.Path,
			Model:     chatReq.Model,
			Tokens:    tokens,
			Duration:  duration,
			Success:   resp.StatusCode == http.StatusOK,
			CreatedAt: time.Now(),
		})
	}()

	return nil
}

// ChatResponse represents an OpenAI-compatible chat response
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// countingReader counts bytes read
type countingReader struct {
	reader io.Reader
	count  int
}

func (cr *countingReader) Read(p []byte) (int, error) {
	n, err := cr.reader.Read(p)
	cr.count += n
	return n, err
}
