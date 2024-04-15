package data

import (
	"context"
	"pt/internal/biz/inter"
	"pt/internal/biz/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
)

type HitRuns struct {
	data *Data
	log  *log.Helper
}

func NewHitRuns(data *Data, logger log.Logger) inter.HitRunsRepo {
	return &HitRuns{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (o *HitRuns) Create(ctx context.Context, hr *model.HitRuns) error {
	err := o.data.DB.WithContext(ctx).Create(hr).Error
	if err != nil {
		return errors.Wrap(err, "Create")
	}
	return nil
}
