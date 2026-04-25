package feed

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/feedfarmer/feedfarmer/internal/storage"
	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
)

type Fetcher struct {
	db *storage.DB
}

func NewFetcher(db *storage.DB) *Fetcher {
	return &Fetcher{db: db}
}

func (f *Fetcher) Fetch(ctx context.Context, feedID, url string) error {
	parser := gofeed.NewParser()
	parsed, err := parser.ParseURLWithContext(url, ctx)
	if err != nil {
		return fmt.Errorf("parse %s: %w", url, err)
	}

	for _, entry := range parsed.Items {
		if entry.Link == "" {
			continue
		}

		var publishedAt *time.Time
		if t := entry.PublishedParsed; t != nil {
			publishedAt = t
		} else if t := entry.UpdatedParsed; t != nil {
			publishedAt = t
		}

		content := entry.Content
		if content == "" {
			content = entry.Description
		}

		item := &storage.Item{
			ID:          uuid.New().String(),
			FeedID:      feedID,
			Title:       entry.Title,
			Link:        entry.Link,
			Content:     content,
			PublishedAt: publishedAt,
			AITags:      []string{},
		}
		if err := f.db.UpsertItem(item); err != nil {
			log.Printf("upsert %s: %v", entry.Link, err)
		}
	}

	return f.db.UpdateFeedMetadata(feedID, parsed.Title, parsed.Description, time.Now())
}
