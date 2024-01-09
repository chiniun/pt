package data

import (
	"context"
	"pt/internal/biz/model"

	//"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
)

type Torrent struct {
	data *Data
	log  *log.Helper
}

func NewTorrent(data *Data, logger log.Logger) *Torrent {
	return &Torrent{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (o *Torrent) Create(ctx context.Context, Torrent *model.Torrent) error {
	err := o.data.DB.WithContext(ctx).Create(Torrent).Error
	if err != nil {
		return errors.Wrap(err,"Create")
	}

	return nil
}
