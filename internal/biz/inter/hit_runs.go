package inter

import (
	"context"
	"pt/internal/biz/model"
)

type HitRunsRepo interface {
	Create(ctx context.Context, hr *model.HitRuns) error
}
