package api

import (
	"encoding/json"
	"net/http"

	"github.com/feedfarmer/feedfarmer/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *handler) listFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := h.db.ListFeeds()
	if err != nil {
		httpError(w, err, 500)
		return
	}
	if feeds == nil {
		feeds = []*storage.Feed{}
	}
	jsonOK(w, feeds)
}

func (h *handler) createFeed(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL                  string `json:"url"`
		FetchIntervalMinutes int    `json:"fetch_interval_minutes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}
	if req.FetchIntervalMinutes <= 0 {
		req.FetchIntervalMinutes = 60
	}

	f := &storage.Feed{
		ID:                   uuid.New().String(),
		URL:                  req.URL,
		FetchIntervalMinutes: req.FetchIntervalMinutes,
	}
	if err := h.db.CreateFeed(f); err != nil {
		httpError(w, err, 500)
		return
	}

	// Async initial fetch
	go func() {
		if err := h.scheduler.RefreshFeed(f.ID, f.URL); err != nil {
			// logged inside RefreshFeed
			_ = err
		}
	}()

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, f)
}

func (h *handler) deleteFeed(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.db.DeleteFeed(id); err != nil {
		httpError(w, err, 500)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) refreshFeed(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	f, err := h.db.GetFeed(id)
	if err != nil || f == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	go h.scheduler.RefreshFeed(f.ID, f.URL)
	jsonOK(w, map[string]string{"status": "refreshing"})
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func httpError(w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), code)
}
