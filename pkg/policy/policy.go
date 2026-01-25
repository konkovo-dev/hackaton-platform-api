package policy

import (
	"context"
)

type Policy[TParams any] interface {
	Action() Action

	LoadContext(ctx context.Context, params TParams) (PolicyContext, error)

	Check(ctx context.Context, pctx PolicyContext) *Decision
}

type BasePolicy struct {
	action Action
}

func NewBasePolicy(action Action) BasePolicy {
	return BasePolicy{action: action}
}

func (p *BasePolicy) Action() Action {
	return p.action
}
