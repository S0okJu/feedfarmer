package storage

import (
	"encoding/json"
	"time"
)

type Item struct {
	ID           string     `json:"id"`
	FeedID       string     `json:"feed_id"`
	FeedTitle    string     `json:"feed_title,omitempty"`
	Title        string     `json:"title"`
	Link         string     `json:"link"`
	Content      string     `json:"content"`
	PublishedAt  *time.Time `json:"published_at"`
	AISummary    string     `json:"ai_summary"`
	AITags       []string   `json:"ai_tags"`
	AIScore      float64    `json:"ai_score"`
	IsRead       bool       `json:"is_read"`
	IsBookmarked bool       `json:"is_bookmarked"`
	CreatedAt    time.Time  `json:"created_at"`
}

// UpsertItem inserts a new item; returns true if a new row was inserted (false = duplicate).
func (db *DB) UpsertItem(item *Item) (bool, error) {
	result, err := db.Exec(
		`INSERT INTO items (id, feed_id, title, link, content, published_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(link) DO NOTHING`,
		item.ID, item.FeedID, item.Title, item.Link, item.Content, item.PublishedAt,
	)
	if err != nil {
		return false, err
	}
	n, _ := result.RowsAffected()
	return n > 0, nil
}

type ItemFilter struct {
	FeedID       string
	OnlyUnread   bool
	OnlyBookmark bool
	Search       string
	Limit        int
	Offset       int
}

func (db *DB) ListItems(f ItemFilter) ([]*Item, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}

	query := `
		SELECT i.id, i.feed_id, fd.title AS feed_title,
		       i.title, i.link, i.content, i.published_at,
		       i.ai_summary, i.ai_tags, i.ai_score,
		       i.is_read, i.is_bookmarked, i.created_at
		FROM items i
		JOIN feeds fd ON fd.id = i.feed_id
		WHERE 1=1`
	args := []any{}

	if f.FeedID != "" {
		query += ` AND i.feed_id = ?`
		args = append(args, f.FeedID)
	}
	if f.OnlyUnread {
		query += ` AND i.is_read = 0`
	}
	if f.OnlyBookmark {
		query += ` AND i.is_bookmarked = 1`
	}
	if f.Search != "" {
		query += ` AND (i.title LIKE ? OR i.content LIKE ?)`
		args = append(args, "%"+f.Search+"%", "%"+f.Search+"%")
	}

	query += ` ORDER BY COALESCE(i.published_at, i.created_at) DESC LIMIT ? OFFSET ?`
	args = append(args, f.Limit, f.Offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		item := &Item{}
		var tagsJSON string
		if err := rows.Scan(
			&item.ID, &item.FeedID, &item.FeedTitle,
			&item.Title, &item.Link, &item.Content, &item.PublishedAt,
			&item.AISummary, &tagsJSON, &item.AIScore,
			&item.IsRead, &item.IsBookmarked, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(tagsJSON), &item.AITags); err != nil {
			item.AITags = []string{}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (db *DB) GetItem(id string) (*Item, error) {
	item := &Item{}
	var tagsJSON string
	err := db.QueryRow(`
		SELECT i.id, i.feed_id, fd.title,
		       i.title, i.link, i.content, i.published_at,
		       i.ai_summary, i.ai_tags, i.ai_score,
		       i.is_read, i.is_bookmarked, i.created_at
		FROM items i
		JOIN feeds fd ON fd.id = i.feed_id
		WHERE i.id = ?`, id,
	).Scan(
		&item.ID, &item.FeedID, &item.FeedTitle,
		&item.Title, &item.Link, &item.Content, &item.PublishedAt,
		&item.AISummary, &tagsJSON, &item.AIScore,
		&item.IsRead, &item.IsBookmarked, &item.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(tagsJSON), &item.AITags); err != nil {
		item.AITags = []string{}
	}
	return item, nil
}

func (db *DB) UpdateItem(id string, isRead, isBookmarked bool) error {
	_, err := db.Exec(
		`UPDATE items SET is_read = ?, is_bookmarked = ? WHERE id = ?`,
		isRead, isBookmarked, id,
	)
	return err
}

// UpdateItemSummary saves an AI-generated summary for an item.
func (db *DB) UpdateItemSummary(id, summary string) error {
	_, err := db.Exec(`UPDATE items SET ai_summary = ? WHERE id = ?`, summary, id)
	return err
}

// UpdateItemAI saves AI-generated tags for an item.
func (db *DB) UpdateItemAI(id string, tags []string) error {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE items SET ai_tags = ? WHERE id = ?`, string(tagsJSON), id)
	return err
}

