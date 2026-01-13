package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Processor struct {
	repo     EventRepository
	handlers map[string]Handler
	cfg      *Config
	logger   *slog.Logger
}

func NewProcessor(repo EventRepository, cfg *Config, logger *slog.Logger) *Processor {
	return &Processor{
		repo:     repo,
		handlers: make(map[string]Handler),
		cfg:      cfg,
		logger:   logger,
	}
}

func (p *Processor) RegisterHandler(handler Handler) {
	p.handlers[handler.EventType()] = handler
	p.logger.Info("handler registered", slog.String("event_type", handler.EventType()))
}

func (p *Processor) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.cfg.PollingInterval)
	defer ticker.Stop()

	p.logger.Info("outbox processor started",
		slog.Duration("polling_interval", p.cfg.PollingInterval),
		slog.Int("batch_size", p.cfg.BatchSize),
		slog.Duration("process_timeout", p.cfg.ProcessTimeout),
		slog.Int("max_attempts", p.cfg.MaxAttempts),
	)

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("outbox processor stopped")
			return nil
		case <-ticker.C:
			if err := p.processBatch(ctx); err != nil {
				p.logger.Error("failed to process batch",
					slog.String("error", err.Error()))
			}
		}
	}
}

func (p *Processor) processBatch(ctx context.Context) error {
	events, err := p.repo.GetPending(ctx, p.cfg.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to get pending events: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	p.logger.Debug("processing events batch", slog.Int("count", len(events)))

	for _, event := range events {
		procCtx, cancel := context.WithTimeout(ctx, p.cfg.ProcessTimeout)

		if err := p.processSingle(procCtx, event); err != nil {
			p.logger.Error("failed to process event",
				slog.String("event_id", event.ID.String()),
				slog.String("event_type", event.EventType),
				slog.String("aggregate_id", event.AggregateID),
				slog.Int("attempt_count", event.AttemptCount),
				slog.String("error", err.Error()),
			)
		}

		cancel()
	}

	return nil
}

func (p *Processor) processSingle(ctx context.Context, event *Event) error {
	event.Status = EventStatusProcessing
	if err := p.repo.Update(ctx, event); err != nil {
		return fmt.Errorf("failed to mark event as processing: %w", err)
	}

	handler, ok := p.handlers[event.EventType]
	if !ok {
		return fmt.Errorf("handler not found for event type: %s", event.EventType)
	}

	event.AttemptCount++
	if err := handler.Handle(ctx, event); err != nil {
		event.LastError = err.Error()

		if event.AttemptCount >= p.cfg.MaxAttempts {
			event.Status = EventStatusFailed
		} else {
			event.Status = EventStatusPending
		}
	} else {
		event.Status = EventStatusProcessed
	}

	updateCtx := context.Background()
	if err := p.repo.Update(updateCtx, event); err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	return nil
}
