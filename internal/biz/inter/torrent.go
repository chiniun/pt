package inter

import (
	"context"
	"pt/internal/biz/model"
)

type TorrentRepo interface {
	FindByHash(ctx context.Context, infoHash string) (views *model.TorrentView, err error)
	UpdateByMap(ctx context.Context, id int64, info map[string]interface{}) error
	GetBuyLogs(ctx context.Context, torrentId int64) ([]*model.TorrentBuyLog, error)
	CreateBuyLog(ctx context.Context, log *model.TorrentBuyLog) error
	Get(ctx context.Context, id int64) (*model.Torrent, error)
}
