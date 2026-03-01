package sync

import (
	"context"
	"log/slog"

	natsclient "github.com/belikoooova/hackaton-platform-api/pkg/nats"
	"github.com/nats-io/nats.go"
	"go.uber.org/fx"
)

type NATSHandler interface {
	Subject() string
	Handle(ctx context.Context, msg *nats.Msg) error
}

var Module = fx.Module("sync",
	fx.Provide(
		fx.Annotate(
			NewUserSkillsUpdatedHandler,
			fx.As(new(NATSHandler)),
			fx.ResultTags(`group:"nats_handlers"`),
		),
		fx.Annotate(
			NewTeamCreatedHandler,
			fx.As(new(NATSHandler)),
			fx.ResultTags(`group:"nats_handlers"`),
		),
		fx.Annotate(
			NewTeamUpdatedHandler,
			fx.As(new(NATSHandler)),
			fx.ResultTags(`group:"nats_handlers"`),
		),
		fx.Annotate(
			NewTeamDeletedHandler,
			fx.As(new(NATSHandler)),
			fx.ResultTags(`group:"nats_handlers"`),
		),
		fx.Annotate(
			NewVacancyCreatedHandler,
			fx.As(new(NATSHandler)),
			fx.ResultTags(`group:"nats_handlers"`),
		),
		fx.Annotate(
			NewVacancyUpdatedHandler,
			fx.As(new(NATSHandler)),
			fx.ResultTags(`group:"nats_handlers"`),
		),
		fx.Annotate(
			NewParticipationRegisteredHandler,
			fx.As(new(NATSHandler)),
			fx.ResultTags(`group:"nats_handlers"`),
		),
		fx.Annotate(
			NewParticipationUpdatedHandler,
			fx.As(new(NATSHandler)),
			fx.ResultTags(`group:"nats_handlers"`),
		),
		fx.Annotate(
			NewParticipationStatusChangedHandler,
			fx.As(new(NATSHandler)),
			fx.ResultTags(`group:"nats_handlers"`),
		),
		fx.Annotate(
			NewParticipationTeamAssignedHandler,
			fx.As(new(NATSHandler)),
			fx.ResultTags(`group:"nats_handlers"`),
		),
		fx.Annotate(
			NewParticipationTeamRemovedHandler,
			fx.As(new(NATSHandler)),
			fx.ResultTags(`group:"nats_handlers"`),
		),
	),
	fx.Invoke(
		fx.Annotate(
			func(
				lc fx.Lifecycle,
				client *natsclient.Client,
				logger *slog.Logger,
				handlers []NATSHandler,
			) error {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				for _, handler := range handlers {
					h := handler
					subject := h.Subject()

					_, err := client.Subscribe(subject, func(msg *nats.Msg) {
						if err := h.Handle(context.Background(), msg); err != nil {
							logger.Error("failed to handle NATS message",
								"subject", subject,
								"error", err,
							)
						}
					})

					if err != nil {
						return err
					}

					logger.Info("subscribed to NATS subject",
						"subject", subject,
					)
				}
				return nil
			},
		})

		return nil
			},
			fx.ParamTags(``, ``, ``, `group:"nats_handlers"`),
		),
	),
)
