package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/feedfarmer/feedfarmer/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *handler) listAIConfigs(w http.ResponseWriter, r *http.Request) {
	configs, err := h.db.ListAIConfigs()
	if err != nil {
		httpError(w, err, 500)
		return
	}
	if configs == nil {
		configs = []*storage.AIConfig{}
	}
	jsonOK(w, configs)
}

func (h *handler) createAIConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Provider string `json:"provider"`
		BaseURL  string `json:"base_url"`
		Model    string `json:"model"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BaseURL == "" {
		http.Error(w, "base_url required", http.StatusBadRequest)
		return
	}
	if req.Provider == "" {
		req.Provider = "ollama"
	}

	cfg := &storage.AIConfig{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Provider: req.Provider,
		BaseURL:  req.BaseURL,
		Model:    req.Model,
		IsActive: req.IsActive,
	}
	if err := h.db.CreateAIConfig(cfg); err != nil {
		httpError(w, err, 500)
		return
	}
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, cfg)
}

func (h *handler) deleteAIConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.db.DeleteAIConfig(id); err != nil {
		httpError(w, err, 500)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) activateAIConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.db.SetActiveAIConfig(id); err != nil {
		httpError(w, err, 500)
		return
	}
	jsonOK(w, map[string]string{"status": "activated"})
}

// tagItem manually triggers AI tagging for a single item.
func (h *handler) tagItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := h.db.GetItem(id)
	if err != nil || item == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	go func() {
		if err := h.aiMgr.TagItem(context.Background(), item.ID, item.Title, item.Content); err != nil {
			// logged inside TagItem; nothing more to do
			_ = err
		}
	}()
	jsonOK(w, map[string]string{"status": "tagging"})
}
