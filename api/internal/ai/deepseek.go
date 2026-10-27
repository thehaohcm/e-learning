// Package ai talks to an OpenAI-compatible chat API (DeepSeek by default) to
// generate cards and to evaluate writing.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var ErrNotConfigured = errors.New("AI is not configured (set DEEPSEEK_API_KEY)")

type Client struct {
	apiKey  string
	baseURL string
	model   string
	http    *http.Client

	mu        sync.Mutex
	modelList []string  // cached /v1/models ids
	modelAt   time.Time // when the list was fetched
}

func NewClient(apiKey, baseURL, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		http:    &http.Client{Timeout: 180 * time.Second},
	}
}

func (c *Client) Enabled() bool { return c != nil && c.apiKey != "" }
func (c *Client) Model() string { return c.model }

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string            `json:"model"`
	Messages       []chatMessage     `json:"messages"`
	Temperature    float64           `json:"temperature"`
	MaxTokens      int               `json:"max_tokens"`
	Stream         bool              `json:"stream"` // always false: we want one JSON response
	ResponseFormat map[string]string `json:"response_format,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

// chatJSON sends a prompt that must be answered with a JSON object and
// decodes that object into out. Works with OpenAI-compatible APIs and
// routers (DeepSeek, 9Router, OpenRouter, ...):
//   - if the API rejects response_format, it retries once without it;
//   - if the API streams anyway (SSE "data: ..." lines), the chunks are joined.
func (c *Client) chatJSON(ctx context.Context, system, user string, temperature float64, out any) error {
	return c.chatJSONModel(ctx, "", system, user, temperature, out)
}

// chatJSONModel is chatJSON with an explicit model ("" = the default model).
func (c *Client) chatJSONModel(ctx context.Context, model, system, user string, temperature float64, out any) error {
	if !c.Enabled() {
		return ErrNotConfigured
	}
	if model == "" {
		model = c.model
	}
	req := chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature:    temperature,
		MaxTokens:      8000,
		ResponseFormat: map[string]string{"type": "json_object"},
	}
	text, status, err := c.complete(ctx, req)
	if err != nil && status == http.StatusBadRequest {
		req.ResponseFormat = nil // some models/routers do not support JSON mode
		text, _, err = c.complete(ctx, req)
	}
	if err != nil {
		return err
	}
	content := extractJSON(text)
	if err := json.Unmarshal([]byte(content), out); err != nil {
		return fmt.Errorf("AI returned invalid JSON: %w (start: %q)", err, truncate(content, 120))
	}
	return nil
}

// complete performs one chat completion and returns the assistant text.
func (c *Client) complete(ctx context.Context, cr chatRequest) (string, int, error) {
	body, err := json.Marshal(cr)
	if err != nil {
		return "", 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("call AI api: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return "", resp.StatusCode, err
	}
	if resp.StatusCode != http.StatusOK {
		return "", resp.StatusCode, fmt.Errorf("AI api error (status %d): %s", resp.StatusCode, truncate(string(raw), 300))
	}

	trimmed := bytes.TrimSpace(raw)
	if bytes.HasPrefix(trimmed, []byte("data:")) || strings.Contains(resp.Header.Get("Content-Type"), "event-stream") {
		text, err := joinSSE(trimmed)
		return text, resp.StatusCode, err
	}
	var out chatResponse
	if err := json.Unmarshal(trimmed, &out); err != nil {
		return "", resp.StatusCode, fmt.Errorf("decode AI response: %w (start: %q)", err, truncate(string(trimmed), 120))
	}
	if len(out.Choices) == 0 {
		return "", resp.StatusCode, errors.New("AI returned no choices")
	}
	return out.Choices[0].Message.Content, resp.StatusCode, nil
}

// joinSSE rebuilds the full answer from a streamed (server-sent events) response:
//
//	data: {"choices":[{"delta":{"content":"Hel"}}]}
//	data: {"choices":[{"delta":{"content":"lo"}}]}
//	data: [DONE]
func joinSSE(raw []byte) (string, error) {
	var b strings.Builder
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		var chunk struct {
			Choices []struct {
				Delta   chatMessage `json:"delta"`
				Message chatMessage `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue // ignore keep-alive or vendor-specific lines
		}
		for _, ch := range chunk.Choices {
			b.WriteString(ch.Delta.Content)
			b.WriteString(ch.Message.Content)
		}
	}
	if b.Len() == 0 {
		return "", errors.New("AI stream contained no content")
	}
	return b.String(), nil
}

// extractJSON strips markdown code fences and any text around the outermost JSON object.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "```"); i >= 0 {
		s = s[i+3:]
		s = strings.TrimPrefix(s, "json")
		if j := strings.Index(s, "```"); j >= 0 {
			s = s[:j]
		}
	}
	start, end := strings.IndexAny(s, "{["), strings.LastIndexAny(s, "}]")
	if start >= 0 && end > start {
		s = s[start : end+1]
	}
	return strings.TrimSpace(s)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// modelsResponse is the OpenAI-compatible /v1/models shape.
type modelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// ListModels returns the model ids the endpoint serves (GET /models), cached
// for a few minutes. Empty (no error) when the endpoint does not list models.
func (c *Client) ListModels(ctx context.Context) ([]string, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	c.mu.Lock()
	if time.Since(c.modelAt) < 5*time.Minute && c.modelList != nil {
		ids := c.modelList
		c.mu.Unlock()
		return ids, nil
	}
	c.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("models: HTTP %d", resp.StatusCode)
	}
	var mr modelsResponse
	if err := json.Unmarshal(raw, &mr); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(mr.Data))
	for _, m := range mr.Data {
		if id := strings.TrimSpace(m.ID); id != "" {
			ids = append(ids, id)
		}
	}
	c.mu.Lock()
	c.modelList = ids
	c.modelAt = time.Now()
	c.mu.Unlock()
	return ids, nil
}
