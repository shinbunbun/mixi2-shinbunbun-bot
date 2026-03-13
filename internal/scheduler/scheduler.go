package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/shinbunbun/mixi2-shinbunbun-bot/internal/github"
	"github.com/shinbunbun/mixi2-shinbunbun-bot/internal/mixi2"
	"github.com/shinbunbun/mixi2-shinbunbun-bot/internal/summary"
)

type Scheduler struct {
	cron         *cron.Cron
	githubClient *github.Client
	mixi2Client  *mixi2.Client
	logger       *slog.Logger
}

func New(githubClient *github.Client, mixi2Client *mixi2.Client) *Scheduler {
	return &Scheduler{
		cron:         cron.New(),
		githubClient: githubClient,
		mixi2Client:  mixi2Client,
		logger:       slog.Default(),
	}
}

func (s *Scheduler) Start(cronExpr string) error {
	_, err := s.cron.AddFunc(cronExpr, func() {
		s.postDailySummary()
	})
	if err != nil {
		return err
	}

	s.cron.Start()
	s.logger.Info("scheduler started", slog.String("cron", cronExpr))
	return nil
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("scheduler stopped")
}

func (s *Scheduler) postDailySummary() {
	s.logger.Info("running daily summary job")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	since := time.Now().Add(-24 * time.Hour)
	events, err := s.githubClient.FetchRecentEvents(ctx, since)
	if err != nil {
		s.logger.Error("failed to fetch github events", slog.String("error", err.Error()))
		return
	}

	text := summary.Generate(events)
	s.logger.Info("generated summary", slog.String("text", text))

	if err := s.mixi2Client.CreatePost(ctx, text); err != nil {
		s.logger.Error("failed to create post", slog.String("error", err.Error()))
		return
	}

	s.logger.Info("daily summary posted successfully")
}
