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

func (o *HitRuns) Get(ctx context.Context, tid, uid int64) (*model.HitRuns, error) {

	hr := new(model.HitRuns)
	err := o.data.DB.WithContext(ctx).Where("torrent_id = ?", tid).Where("user_id", uid).First(hr).Error
	if err != nil {
		return nil, errors.Wrap(err, "Create")
	}
	return hr, nil
}
