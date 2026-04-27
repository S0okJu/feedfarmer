package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// OllamaProvider works with Ollama-compatible servers, including custom Gemma deployments.
// Expected API: POST /chat with {"messages":[{"role":"user","content":"..."}]}
type OllamaProvider struct {
	BaseURL string // e.g. "http://192.168.50.27:30080"
	Model   string // informational only — server selects model
	Logger  *slog.Logger
	Client  *http.Client
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

	result, err := p.chat(ctx, "ai.ollama.tag", prompt)
	if err != nil {
		return nil, err
	}

	raw := strings.TrimSpace(result.Message.Content)
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start == -1 || end <= start {
		aiErr := &Error{
			Kind:         ErrInvalidResponse,
			Op:           "ai.ollama.tag.invalidResponse",
			PublicReason: fmt.Sprintf("model response format invalid: %.100s", raw),
		}
		p.logAIError(ctx, "ollama tag failed", aiErr)
		return nil, aiErr
	}

	var tags []string
	if err := json.Unmarshal([]byte(raw[start:end+1]), &tags); err != nil {
		aiErr := &Error{
			Kind:         ErrResponseParse,
			Op:           "ai.ollama.tag.parseTags",
			PublicReason: "failed to parse tags from model response",
			Err:          err,
		}
		p.logAIError(ctx, "ollama tag failed", aiErr)
		return nil, aiErr
	}
	p.logger().InfoContext(ctx, "ollama tag succeeded", p.opAttrs(ctx, "ai.ollama.tag", len(prompt), "tags_count", len(tags))...)
	return tags, nil
}

func (p *OllamaProvider) Summarize(ctx context.Context, content string) (string, error) {
	prompt := fmt.Sprintf(
		"Summarize the following article in 2-3 sentences. Be concise and focus on key points.\n"+
			"Return ONLY the summary text, no labels or explanation.\n\n%s",
		truncate(content, 4000),
	)

	result, err := p.chat(ctx, "ai.ollama.summarize", prompt)
	if err != nil {
		return "", err
	}

	summary := strings.TrimSpace(result.Message.Content)
	if summary == "" {
		aiErr := &Error{
			Kind:         ErrInvalidResponse,
			Op:           "ai.ollama.summarize.emptyResponse",
			PublicReason: "model returned empty summary",
		}
		p.logAIError(ctx, "ollama summarize failed", aiErr)
		return "", aiErr
	}

	p.logger().InfoContext(ctx, "ollama summarize succeeded", p.opAttrs(ctx, "ai.ollama.summarize", len(prompt), "summary_chars", len(summary))...)
	return summary, nil
}

func (p *OllamaProvider) chat(ctx context.Context, op, prompt string) (ollamaChatResponse, error) {
	logger := p.logger()
	started := time.Now()
	baseAttrs := p.opAttrs(ctx, op, len(prompt))
	logger.InfoContext(ctx, "ollama request started", baseAttrs...)

	reqBody, err := json.Marshal(ollamaChatRequest{
		Messages: []ollamaMsg{{Role: "user", Content: prompt}},
	})
	if err != nil {
		aiErr := &Error{
			Kind:         ErrRequestBuild,
			Op:           op + ".marshalRequest",
			PublicReason: "failed to encode model request",
			Err:          err,
		}
		p.logAIError(ctx, "ollama request failed", aiErr, "latency_ms", time.Since(started).Milliseconds())
		return ollamaChatResponse{}, aiErr
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/chat", bytes.NewReader(reqBody))
	if err != nil {
		aiErr := &Error{
			Kind:         ErrRequestBuild,
			Op:           op + ".newRequest",
			PublicReason: "failed to build model request",
			Err:          err,
		}
		p.logAIError(ctx, "ollama request failed", aiErr, "latency_ms", time.Since(started).Milliseconds())
		return ollamaChatResponse{}, aiErr
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client().Do(req)
	if err != nil {
		aiErr := &Error{
			Kind:         ErrRequestFailed,
			Op:           op + ".doRequest",
			PublicReason: "failed to contact model server",
			Err:          err,
		}
		p.logAIError(ctx, "ollama request failed", aiErr, "latency_ms", time.Since(started).Milliseconds())
		return ollamaChatResponse{}, aiErr
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		aiErr := &Error{
			Kind:         ErrResponseParse,
			Op:           op + ".readResponse",
			PublicReason: "failed to read model response",
			Err:          err,
		}
		p.logAIError(
			ctx,
			"ollama request failed",
			aiErr,
			"latency_ms", time.Since(started).Milliseconds(),
			"status_code", resp.StatusCode,
		)
		return ollamaChatResponse{}, aiErr
	}

	if resp.StatusCode >= 400 {
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		aiErr := &Error{
			Kind:         ErrUpstreamStatus,
			Op:           op + ".upstreamStatus",
			StatusCode:   resp.StatusCode,
			PublicReason: fmt.Sprintf("model server returned %d: %s", resp.StatusCode, clip(msg, 512)),
		}
		p.logAIError(
			ctx,
			"ollama request failed",
			aiErr,
			"latency_ms", time.Since(started).Milliseconds(),
			"response_bytes", len(body),
		)
		return ollamaChatResponse{}, aiErr
	}

	var result ollamaChatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		aiErr := &Error{
			Kind:         ErrResponseParse,
			Op:           op + ".decodeResponse",
			PublicReason: "failed to parse model response",
			Err:          err,
		}
		p.logAIError(
			ctx,
			"ollama request failed",
			aiErr,
			"latency_ms", time.Since(started).Milliseconds(),
			"status_code", resp.StatusCode,
			"response_bytes", len(body),
		)
		return ollamaChatResponse{}, aiErr
	}

	logger.InfoContext(
		ctx,
		"ollama request succeeded",
		p.opAttrs(
			ctx,
			op,
			len(prompt),
			"status_code", resp.StatusCode,
			"latency_ms", time.Since(started).Milliseconds(),
			"response_bytes", len(body),
		)...,
	)
	return result, nil
}

func (p *OllamaProvider) logger() *slog.Logger {
	if p.Logger != nil {
		return p.Logger
	}
	return slog.Default()
}

func (p *OllamaProvider) client() *http.Client {
	if p.Client != nil {
		return p.Client
	}
	return http.DefaultClient
}

func (p *OllamaProvider) logAIError(ctx context.Context, message string, aiErr *Error, extra ...any) {
	attrs := p.opAttrs(ctx, "", -1)
	attrs = append(attrs, aiErr.LogAttrs()...)
	attrs = append(attrs, extra...)
	p.logger().ErrorContext(ctx, message, attrs...)
}

func (p *OllamaProvider) opAttrs(ctx context.Context, op string, promptChars int, extra ...any) []any {
	attrs := []any{"provider", "ollama", "model", p.Model, "base_url", p.BaseURL}
	if op != "" {
		attrs = append(attrs, "op", op)
	}
	if promptChars >= 0 {
		attrs = append(attrs, "prompt_chars", promptChars)
	}
	if requestID := middleware.GetReqID(ctx); requestID != "" {
		attrs = append(attrs, "request_id", requestID)
	}
	attrs = append(attrs, extra...)
	return attrs
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
