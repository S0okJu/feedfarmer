package ai

import (
	"context"
	"fmt"
	"time"

	readability "github.com/go-shiori/go-readability"

	"github.com/feedfarmer/feedfarmer/internal/storage"
)

// Manager loads the active AI config from DB and dispatches tagging jobs.
type Manager struct {
	db *storage.DB
}

func NewManager(db *storage.DB) *Manager {
	return &Manager{db: db}
}

func (m *Manager) activeProvider() (Provider, error) {
	cfg, err := m.db.GetActiveAIConfig()
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, fmt.Errorf("no active AI config")
	}
	return &OllamaProvider{BaseURL: cfg.BaseURL, Model: cfg.Model}, nil
}

// SummarizeItem fetches the article at link, extracts its text, summarizes it via AI, and saves the result.
func (m *Manager) SummarizeItem(ctx context.Context, itemID, link string) (string, error) {
	provider, err := m.activeProvider()
	if err != nil {
		return "", err
	}

	article, err := readability.FromURL(link, 30*time.Second)
	if err != nil {
		return "", fmt.Errorf("fetch article: %w", err)
	}

	content := article.TextContent
	if content == "" {
		content = article.Excerpt
	}

	summary, err := provider.Summarize(ctx, content)
	if err != nil {
		return "", err
	}

	if err := m.db.UpdateItemSummary(itemID, summary); err != nil {
		return "", err
	}
	return summary, nil
}

// TagItem runs AI tagging for an item and persists the result.
func (m *Manager) TagItem(ctx context.Context, itemID, title, content string) error {
	provider, err := m.activeProvider()
	if err != nil {
		return err
	}
	tags, err := provider.Tag(ctx, title, content)
	if err != nil {
		return err
	}
	return m.db.UpdateItemAI(itemID, tags)
}
