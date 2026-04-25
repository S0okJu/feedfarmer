package feed

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/feedfarmer/feedfarmer/internal/ai"
	"github.com/feedfarmer/feedfarmer/internal/storage"
)

type Scheduler struct {
	db      *storage.DB
	fetcher *Fetcher
	stop    chan struct{}
	wg      sync.WaitGroup
}

func NewScheduler(db *storage.DB, aiMgr *ai.Manager) *Scheduler {
	return &Scheduler{
		db:      db,
		fetcher: NewFetcher(db, aiMgr),
		stop:    make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	s.wg.Add(1)
	go s.run()
}

func (s *Scheduler) Stop() {
	close(s.stop)
	s.wg.Wait()
}

// RefreshFeed manually triggers a fetch for a single feed.
func (s *Scheduler) RefreshFeed(feedID, url string) error {
	return s.fetcher.Fetch(context.Background(), feedID, url)
}

// RefreshAll fetches all subscribed feeds concurrently.
func (s *Scheduler) RefreshAll() {
	feeds, err := s.db.AllFeeds()
	if err != nil {
		log.Printf("scheduler list feeds: %v", err)
		return
	}
	for _, f := range feeds {
		go func(id, url string) {
			if err := s.fetcher.Fetch(context.Background(), id, url); err != nil {
				log.Printf("scheduler fetch %s: %v", url, err)
			}
		}(f.ID, f.URL)
	}
}

func (s *Scheduler) run() {
	defer s.wg.Done()
	// Fetch on startup
	s.RefreshAll()

	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.RefreshAll()
		case <-s.stop:
			return
		}
	}
}
