package inter

import (
	"context"
	"pt/internal/biz/constant"
	"pt/internal/biz/model"
)

type PeerRepo interface {
	GetPeerList(ctx context.Context, tid int64, onlyLeechQuery, limit string) ([]*model.PeerView, error)
	GetPeerView(ctx context.Context, tid, uid int64, peer_id string) (*model.PeerView, error)
	GetPeer(ctx context.Context, tid, uid int64, ip, peerId string) (*model.Peer, error)
	GetPeerListByUser(ctx context.Context, tid, uid int64) ([]*model.Peer, error)
	CountPeersByUserAndSeedType(ctx context.Context, uid int64, seederType string) (int64, error)
	Delete(ctx context.Context, id int64) (int64, error)
	Update(ctx context.Context, id int64, updateMap map[string]interface{}) (*model.Peer, error)
	Insert(ctx context.Context, peer *model.Peer) error
	GetCountByTrackerState(ctx context.Context, tid int64, state constant.TrackerState) (int64, error)
}
