package scraper

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

// Scheduler menjalankan loop yang mengecek tiap menit source mana yang
// sudah waktunya di-scrape berdasarkan cron expression masing-masing.
type Scheduler struct {
	sourceRepo SourceRepo
	pipeline   *Pipeline
	parser     cron.Parser
}

// NewScheduler membuat scheduler baru.
func NewScheduler(repo SourceRepo, pipe *Pipeline) *Scheduler {
	return &Scheduler{
		sourceRepo: repo,
		pipeline:   pipe,
		parser: cron.NewParser(
			cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
		),
	}
}

// Run memulai loop scheduler. Berhenti saat ctx dibatalkan.
func (s *Scheduler) Run(ctx context.Context) {
	slog.Info("news scheduler started")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// jalankan sekali di awal agar tidak menunggu 1 menit pertama
	s.tick(ctx, time.Now())

	for {
		select {
		case <-ctx.Done():
			slog.Info("news scheduler stopped")
			return
		case now := <-ticker.C:
			s.tick(ctx, now)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context, now time.Time) {
	sources, err := s.sourceRepo.ListActive(ctx)
	if err != nil {
		slog.Error("list active sources", "err", err)
		return
	}

	for _, src := range sources {
		if !s.isDue(src, now) {
			continue
		}
		src := src
		go func() {
			slog.Info("scraping source", "source", src.Key)
			if err := s.pipeline.ScrapeSource(ctx, src); err != nil {
				slog.Error("scrape source", "source", src.Key, "err", err)
			}
		}()
	}
}

// isDue mengevaluasi apakah source sudah waktunya di-scrape.
func (s *Scheduler) isDue(src Source, now time.Time) bool {
	sched, err := s.parser.Parse(src.Schedule)
	if err != nil {
		slog.Warn("invalid cron expression", "source", src.Key, "schedule", src.Schedule)
		return false
	}
	if src.LastScrapedAt == nil {
		return true // belum pernah di-scrape
	}
	next := sched.Next(*src.LastScrapedAt)
	return !next.After(now)
}
