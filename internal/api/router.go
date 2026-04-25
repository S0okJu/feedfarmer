package api

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/feedfarmer/feedfarmer/internal/feed"
	"github.com/feedfarmer/feedfarmer/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type handler struct {
	db        *storage.DB
	scheduler *feed.Scheduler
}

func NewRouter(db *storage.DB, scheduler *feed.Scheduler, webFS fs.FS) http.Handler {
	h := &handler{db: db, scheduler: scheduler}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type"},
	}))

	r.Route("/api", func(r chi.Router) {
		r.Route("/feeds", func(r chi.Router) {
			r.Get("/", h.listFeeds)
			r.Post("/", h.createFeed)
			r.Delete("/{id}", h.deleteFeed)
			r.Post("/{id}/refresh", h.refreshFeed)
		})
		r.Route("/items", func(r chi.Router) {
			r.Get("/", h.listItems)
			r.Get("/{id}", h.getItem)
			r.Patch("/{id}", h.updateItem)
		})
	})

	// SPA fallback: serve index.html for unknown paths (React Router handles client routing)
	r.Handle("/*", spaHandler(webFS))

	return r
}

func spaHandler(webFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(webFS))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimLeft(r.URL.Path, "/")
		if path != "" {
			if _, err := webFS.Open(path); err == nil {
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		// Fall back to index.html for SPA client-side routing
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
