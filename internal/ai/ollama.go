package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var result ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	raw := strings.TrimSpace(result.Message.Content)
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start == -1 || end <= start {
		return nil, fmt.Errorf("no JSON array in AI response: %.100s", raw)
	}

	var tags []string
	if err := json.Unmarshal([]byte(raw[start:end+1]), &tags); err != nil {
		return nil, fmt.Errorf("parse tags JSON: %w", err)
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
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var result ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return strings.TrimSpace(result.Message.Content), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
