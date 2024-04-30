package inter

import (
	"context"
	"pt/internal/biz/model"
	"time"
)

type CheaterRepo interface {
	Create(ctx context.Context, cheater *model.Cheaters) error
	Get(ctx context.Context, id int64) (*model.Cheaters, error)
	Update(ctx context.Context, cheater *model.Cheaters) error
	Count(ctx context.Context, uid, tid int64, added time.Time) (int64, error)
	Query(ctx context.Context, uid, tid int64, added time.Time) ([]*model.Cheaters, error)
}
