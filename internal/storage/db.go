package storage

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// SQLite is single-writer; WAL allows concurrent reads
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;`); err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS feeds (
			id                     TEXT PRIMARY KEY,
			url                    TEXT NOT NULL UNIQUE,
			title                  TEXT NOT NULL DEFAULT '',
			description            TEXT NOT NULL DEFAULT '',
			last_fetched_at        DATETIME,
			fetch_interval_minutes INTEGER NOT NULL DEFAULT 60,
			created_at             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS items (
			id            TEXT PRIMARY KEY,
			feed_id       TEXT NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
			title         TEXT NOT NULL DEFAULT '',
			link          TEXT NOT NULL UNIQUE,
			content       TEXT NOT NULL DEFAULT '',
			published_at  DATETIME,
			ai_summary    TEXT NOT NULL DEFAULT '',
			ai_tags       TEXT NOT NULL DEFAULT '[]',
			ai_score      REAL NOT NULL DEFAULT 0,
			is_read       INTEGER NOT NULL DEFAULT 0,
			is_bookmarked INTEGER NOT NULL DEFAULT 0,
			created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_items_feed_id   ON items(feed_id)`,
		`CREATE INDEX IF NOT EXISTS idx_items_published ON items(published_at DESC)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
