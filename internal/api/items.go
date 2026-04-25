package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/feedfarmer/feedfarmer/internal/storage"
	"github.com/go-chi/chi/v5"
)

func (h *handler) listItems(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := storage.ItemFilter{
		FeedID:       q.Get("feed_id"),
		OnlyUnread:   q.Get("unread") == "true",
		OnlyBookmark: q.Get("bookmarked") == "true",
		Search:       q.Get("q"),
		Limit:        limit,
		Offset:       offset,
	}

	items, err := h.db.ListItems(filter)
	if err != nil {
		httpError(w, err, 500)
		return
	}
	if items == nil {
		items = []*storage.Item{}
	}
	jsonOK(w, items)
}

func (h *handler) getItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := h.db.GetItem(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	jsonOK(w, item)
}

func (h *handler) updateItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		IsRead       *bool `json:"is_read"`
		IsBookmarked *bool `json:"is_bookmarked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	item, err := h.db.GetItem(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	isRead := item.IsRead
	isBookmarked := item.IsBookmarked
	if req.IsRead != nil {
		isRead = *req.IsRead
	}
	if req.IsBookmarked != nil {
		isBookmarked = *req.IsBookmarked
	}

	if err := h.db.UpdateItem(id, isRead, isBookmarked); err != nil {
		httpError(w, err, 500)
		return
	}
	item.IsRead = isRead
	item.IsBookmarked = isBookmarked
	jsonOK(w, item)
}
