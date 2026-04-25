package storage

import (
	"database/sql"
	"time"
)

type Feed struct {
	ID                   string     `json:"id"`
	URL                  string     `json:"url"`
	Title                string     `json:"title"`
	Description          string     `json:"description"`
	LastFetchedAt        *time.Time `json:"last_fetched_at"`
	FetchIntervalMinutes int        `json:"fetch_interval_minutes"`
	ItemCount            int        `json:"item_count,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

func (db *DB) CreateFeed(f *Feed) error {
	_, err := db.Exec(
		`INSERT INTO feeds (id, url, title, description, fetch_interval_minutes) VALUES (?, ?, ?, ?, ?)`,
		f.ID, f.URL, f.Title, f.Description, f.FetchIntervalMinutes,
	)
	return err
}

func (db *DB) GetFeed(id string) (*Feed, error) {
	f := &Feed{}
	err := db.QueryRow(
		`SELECT id, url, title, description, last_fetched_at, fetch_interval_minutes, created_at
		 FROM feeds WHERE id = ?`, id,
	).Scan(&f.ID, &f.URL, &f.Title, &f.Description, &f.LastFetchedAt, &f.FetchIntervalMinutes, &f.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return f, err
}

func (db *DB) ListFeeds() ([]*Feed, error) {
	rows, err := db.Query(`
		SELECT f.id, f.url, f.title, f.description,
		       f.last_fetched_at, f.fetch_interval_minutes, f.created_at,
		       COUNT(i.id) AS item_count
		FROM feeds f
		LEFT JOIN items i ON i.feed_id = f.id
		GROUP BY f.id
		ORDER BY f.created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []*Feed
	for rows.Next() {
		f := &Feed{}
		if err := rows.Scan(
			&f.ID, &f.URL, &f.Title, &f.Description,
			&f.LastFetchedAt, &f.FetchIntervalMinutes, &f.CreatedAt,
			&f.ItemCount,
		); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}
	return feeds, rows.Err()
}

func (db *DB) DeleteFeed(id string) error {
	_, err := db.Exec(`DELETE FROM feeds WHERE id = ?`, id)
	return err
}

func (db *DB) UpdateFeedMetadata(id, title, description string, fetchedAt time.Time) error {
	_, err := db.Exec(
		`UPDATE feeds SET title = ?, description = ?, last_fetched_at = ? WHERE id = ?`,
		title, description, fetchedAt, id,
	)
	return err
}

// AllFeeds returns minimal feed data used by the scheduler.
func (db *DB) AllFeeds() ([]*Feed, error) {
	rows, err := db.Query(`SELECT id, url, fetch_interval_minutes FROM feeds`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []*Feed
	for rows.Next() {
		f := &Feed{}
		if err := rows.Scan(&f.ID, &f.URL, &f.FetchIntervalMinutes); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}
	return feeds, rows.Err()
}
