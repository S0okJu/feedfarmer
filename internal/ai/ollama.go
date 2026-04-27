package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OllamaProvider works with Ollama-compatible servers, including custom Gemma deployments.
// Expected API: POST /chat with {"messages":[{"role":"user","content":"..."}]}
type OllamaProvider struct {
	BaseURL string // e.g. "http://192.168.50.27:30080"
	Model   string // informational only — server selects model
}

type ollamaChatRequest struct {
	Messages []ollamaMsg `json:"messages"`
}

type ollamaMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatResponse struct {
	Message ollamaMsg `json:"message"`
	Done    bool      `json:"done"`
}

func (p *OllamaProvider) Tag(ctx context.Context, title, text string) ([]string, error) {
	prompt := fmt.Sprintf(
		"Extract 3-7 relevant tags from the following article.\n"+
			"Return ONLY a JSON array of lowercase strings, no explanation.\n"+
			"Example: [\"technology\",\"golang\",\"api\"]\n\n"+
			"Title: %s\nContent: %s",
		title, truncate(text, 1500),
	)

	reqBody, _ := json.Marshal(ollamaChatRequest{
		Messages: []ollamaMsg{{Role: "user", Content: prompt}},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/chat", bytes.NewReader(reqBody))
	if err != nil {
		return nil, &Error{
			Kind:         ErrRequestBuild,
			Op:           "ai.ollama.tag.newRequest",
			PublicReason: "failed to build model request",
			Err:          err,
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &Error{
			Kind:         ErrRequestFailed,
			Op:           "ai.ollama.tag.doRequest",
			PublicReason: "failed to contact model server",
			Err:          err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return nil, &Error{
			Kind:         ErrUpstreamStatus,
			Op:           "ai.ollama.tag.upstreamStatus",
			StatusCode:   resp.StatusCode,
			PublicReason: fmt.Sprintf("model server returned %d: %s", resp.StatusCode, msg),
		}
	}

	var result ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, &Error{
			Kind:         ErrResponseParse,
			Op:           "ai.ollama.tag.decodeResponse",
			PublicReason: "failed to parse model response",
			Err:          err,
		}
	}

	raw := strings.TrimSpace(result.Message.Content)
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start == -1 || end <= start {
		return nil, &Error{
			Kind:         ErrInvalidResponse,
			Op:           "ai.ollama.tag.invalidResponse",
			PublicReason: fmt.Sprintf("model response format invalid: %.100s", raw),
		}
	}

	var tags []string
	if err := json.Unmarshal([]byte(raw[start:end+1]), &tags); err != nil {
		return nil, &Error{
			Kind:         ErrResponseParse,
			Op:           "ai.ollama.tag.parseTags",
			PublicReason: "failed to parse tags from model response",
			Err:          err,
		}
	}
	return tags, nil
}

func (p *OllamaProvider) Summarize(ctx context.Context, content string) (string, error) {
	prompt := fmt.Sprintf(
		"Summarize the following article in 2-3 sentences. Be concise and focus on key points.\n"+
			"Return ONLY the summary text, no labels or explanation.\n\n%s",
		truncate(content, 4000),
	)

	reqBody, _ := json.Marshal(ollamaChatRequest{
		Messages: []ollamaMsg{{Role: "user", Content: prompt}},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/chat", bytes.NewReader(reqBody))
	if err != nil {
		return "", &Error{
			Kind:         ErrRequestBuild,
			Op:           "ai.ollama.summarize.newRequest",
			PublicReason: "failed to build model request",
			Err:          err,
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", &Error{
			Kind:         ErrRequestFailed,
			Op:           "ai.ollama.summarize.doRequest",
			PublicReason: "failed to contact model server",
			Err:          err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return "", &Error{
			Kind:         ErrUpstreamStatus,
			Op:           "ai.ollama.summarize.upstreamStatus",
			StatusCode:   resp.StatusCode,
			PublicReason: fmt.Sprintf("model server returned %d: %s", resp.StatusCode, msg),
		}
	}

	var result ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", &Error{
			Kind:         ErrResponseParse,
			Op:           "ai.ollama.summarize.decodeResponse",
			PublicReason: "failed to parse model response",
			Err:          err,
		}
	}

	summary := strings.TrimSpace(result.Message.Content)
	if summary == "" {
		return "", &Error{
			Kind:         ErrInvalidResponse,
			Op:           "ai.ollama.summarize.emptyResponse",
			PublicReason: "model returned empty summary",
		}
	}

	return summary, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
