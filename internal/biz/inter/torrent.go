package inter

import (
	"context"
	"pt/internal/biz/model"
)

type TorrentRepo interface {
	FindByHash(infoHash string) (views *model.TorrentView, err error)
	GetPeerList(ctx context.Context, tid int64, onlyLeechQuery, limit string) ([]*model.PeerView, error)
}
