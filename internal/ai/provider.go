package ai

import "context"

// Provider abstracts AI model interactions.
type Provider interface {
	Tag(ctx context.Context, title, content string) ([]string, error)
	Summarize(ctx context.Context, content string) (string, error)
}
