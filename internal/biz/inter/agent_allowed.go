package inter

import (
	"context"
	"pt/internal/biz/model"
)

type AgentAllowedRepo interface {
	Create(ctx context.Context, agent *model.AgentAllowedFamily) error
	Get(ctx context.Context) (*model.AgentAllowedFamily, error)
	GetList(ctx context.Context) ([]*model.AgentAllowedFamily, error)
}
