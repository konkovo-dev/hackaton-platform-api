package policy

import (
	"context"
	"log/slog"
)

type Engine struct {
	logger *slog.Logger
}

func NewEngine(logger *slog.Logger) *Engine {
	return &Engine{
		logger: logger,
	}
}

func Enforce[TParams any](ctx context.Context, logger *slog.Logger, policy Policy[TParams], params TParams) error {
	pctx, err := policy.LoadContext(ctx, params)
	if err != nil {
		logger.ErrorContext(ctx, "failed to load policy context",
			slog.String("action", string(policy.Action())),
			slog.String("error", err.Error()),
		)
		return err
	}

	decision := policy.Check(ctx, pctx)

	if !decision.Allowed {
		logger.WarnContext(ctx, "policy check failed",
			slog.String("action", string(policy.Action())),
			slog.Int("violations", len(decision.Violations)),
		)
		for _, v := range decision.Violations {
			logger.WarnContext(ctx, "policy violation",
				slog.String("action", string(policy.Action())),
				slog.String("code", v.Code),
				slog.String("field", v.Field),
				slog.String("message", v.Message),
			)
		}
	} else {
		logger.DebugContext(ctx, "policy check passed",
			slog.String("action", string(policy.Action())),
		)
	}

	if !decision.Allowed {
		return NewPolicyError(policy.Action(), decision.Violations)
	}

	return nil
}

func CheckPolicy[TParams any](ctx context.Context, policy Policy[TParams], params TParams) (*Decision, error) {
	pctx, err := policy.LoadContext(ctx, params)
	if err != nil {
		return nil, err
	}

	return policy.Check(ctx, pctx), nil
}
