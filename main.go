package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/feedfarmer/feedfarmer/internal/ai"
	"github.com/feedfarmer/feedfarmer/internal/api"
	"github.com/feedfarmer/feedfarmer/internal/feed"
	"github.com/feedfarmer/feedfarmer/internal/storage"
)

//go:embed web/dist
var webFS embed.FS

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/feedfarmer.db"
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		log.Fatal("create data dir:", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		log.Fatal("open db:", err)
	}
	defer db.Close()

	aiMgr := ai.NewManager(db)

	scheduler := feed.NewScheduler(db, aiMgr)
	scheduler.Start()
	defer scheduler.Stop()

	distFS, err := fs.Sub(webFS, "web/dist")
	if err != nil {
		log.Fatal("embed fs:", err)
	}

	router := api.NewRouter(db, scheduler, aiMgr, distFS)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("FeedFarmer listening on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
