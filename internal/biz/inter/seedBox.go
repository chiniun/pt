package inter

import (
	"context"
	"pt/internal/biz/model"
)

type SeedBoxRepo interface {
	Get(ctx context.Context, id int64) (*model.SeedBoxRecord, error)
	Update(ctx context.Context, seedBox *model.SeedBoxRecord) error
	Create(ctx context.Context, seedBox *model.SeedBoxRecord) error
	Query(ctx context.Context, seedBox *model.SeedBoxRecord) (*model.SeedBoxRecord, error)
}
