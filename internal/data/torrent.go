package data

import (
	"context"
	"pt/internal/biz"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
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

func (o *Torrent) Create(ctx context.Context, Torrent *biz.Torrent) error {
	return o.data.DB.WithContext(ctx).Create(Torrent).Error

}