package inter

import (
	"context"
	"pt/internal/biz/model"
)

type TorrentRepo interface {
	FindByHash(ctx context.Context, infoHash string) (views *model.TorrentView, err error)
}
