package core

import (
	"fmt"
	"log/slog"

	"github.com/robfig/cron/v3"
)

func (s *Server) RestartCronJob() error {
	if s.State.CronJob != nil {
		s.State.CronJob.Stop()
	}

	if s.AppSettings.CronSchedule == nil || *s.AppSettings.CronSchedule == "" {
		slog.Info("No cron schedule configured, cron job disabled")
		s.State.CronJob = nil
		return nil
	}

	schedule := *s.AppSettings.CronSchedule
	slog.Info("Setting up cron job", "schedule", schedule)

	s.State.CronJob = cron.New()
	_, err := s.State.CronJob.AddFunc(*s.AppSettings.CronSchedule, s.runCronJobOnce)
	if err != nil {
		slog.Error("Failed to add cron job", "error", err)
		return fmt.Errorf("failed to add cron job: %w", err)
	}
	s.State.CronJob.Start()
	return nil
}

func (s *Server) runCronJobOnce() {
	slog.Info("Running scheduled backup")
	s.ProcessAllConfigOptions()
}

func ValidateCronSchedule(schedule string) error {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(schedule)
	return err
}
