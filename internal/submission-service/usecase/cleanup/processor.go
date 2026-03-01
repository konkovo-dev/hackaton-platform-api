package cleanup

import (
	"context"
	"log/slog"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/usecase/submission"
)

type Processor struct {
	fileRepo submission.SubmissionFileRepository
	logger   *slog.Logger
}

func NewProcessor(fileRepo submission.SubmissionFileRepository, logger *slog.Logger) *Processor {
	return &Processor{
		fileRepo: fileRepo,
		logger:   logger,
	}
}

func (p *Processor) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	p.logger.Info("cleanup processor started")

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("cleanup processor stopped")
			return
		case <-ticker.C:
			affected, err := p.fileRepo.MarkExpiredFilesAsFailed(ctx, 3*time.Minute)
			if err != nil {
				p.logger.Error("failed to mark expired files as failed", "error", err)
			} else if affected > 0 {
				p.logger.Info("marked expired files as failed", "count", affected)
			}
		}
	}
}
