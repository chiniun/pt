package inter

import (
	"context"
	"pt/internal/biz/model"
)

type PeerRepo interface {
	GetPeerList(ctx context.Context, tid int64, onlyLeechQuery, limit string) ([]*model.PeerView, error)
	GetPeerView(ctx context.Context, tid, uid int64, peer_id string) (*model.PeerView, error)
	GetPeer(ctx context.Context, tid, uid int64, ip string) (*model.Peer, error)
	GetPeerListByUser(ctx context.Context, tid, uid int64) ([]*model.Peer, error)
}
