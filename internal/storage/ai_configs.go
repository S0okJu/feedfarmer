package storage

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type AIConfig struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Provider  string    `json:"provider"` // "ollama" (more providers in the future)
	BaseURL   string    `json:"base_url"`
	Model     string    `json:"model"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

func (db *DB) CreateAIConfig(cfg *AIConfig) error {
	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
	}
	_, err := db.Exec(
		`INSERT INTO ai_configs (id, name, provider, base_url, model, is_active)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		cfg.ID, cfg.Name, cfg.Provider, cfg.BaseURL, cfg.Model, cfg.IsActive,
	)
	return err
}

func (db *DB) ListAIConfigs() ([]*AIConfig, error) {
	rows, err := db.Query(
		`SELECT id, name, provider, base_url, model, is_active, created_at
		 FROM ai_configs ORDER BY created_at`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*AIConfig
	for rows.Next() {
		c := &AIConfig{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Provider, &c.BaseURL, &c.Model, &c.IsActive, &c.CreatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, rows.Err()
}

func (db *DB) GetActiveAIConfig() (*AIConfig, error) {
	c := &AIConfig{}
	err := db.QueryRow(
		`SELECT id, name, provider, base_url, model, is_active, created_at
		 FROM ai_configs WHERE is_active = 1 ORDER BY created_at LIMIT 1`,
	).Scan(&c.ID, &c.Name, &c.Provider, &c.BaseURL, &c.Model, &c.IsActive, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (db *DB) DeleteAIConfig(id string) error {
	_, err := db.Exec(`DELETE FROM ai_configs WHERE id = ?`, id)
	return err
}

// SetActiveAIConfig deactivates all configs then activates the given one.
func (db *DB) SetActiveAIConfig(id string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE ai_configs SET is_active = 0`); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE ai_configs SET is_active = 1 WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}
