package data

import (
	"context"
	"fmt"
	"pt/internal/biz/model"
	"time"

	//"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
)

type Peer struct {
	data *Data
	log  *log.Helper
}

func NewPeer(data *Data, logger log.Logger) *Peer {
	return &Peer{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (o *Peer) GetPeerList(ctx context.Context, tid int64, onlyLeechQuery, limit string) ([]*model.PeerView, error) {

	peers := make([]*model.PeerView, 0)

	fields := fmt.Sprintf("id, seeder, peer_id, ip, ipv4, ipv6, port, uploaded, downloaded, userid, last_action, UNIX_TIMESTAMP(last_action) as last_action_unix_timestamp, prev_action, (%d - UNIX_TIMESTAMP(last_action)) AS announcetime, UNIX_TIMESTAMP(prev_action) AS prevts", time.Now().Unix())

	peerListSQL := fmt.Sprintf("SELECT %s FROM peers WHERE Peer = %d %s %s", fields, tid, onlyLeechQuery, limit)

	err := o.data.DB.Raw(peerListSQL).Scan(&peers).Error
	if err != nil {
		return nil, errors.Wrap(err, "GetPeerList")
	}
	return peers, nil

}

func (o *Peer) GetPeerView(ctx context.Context, tid, uid int64, peer_id string) (*model.PeerView, error) {
	peer := new(model.PeerView)

	fields := fmt.Sprintf("id, seeder, peer_id, ip, ipv4, ipv6, port, uploaded, downloaded, userid, last_action, UNIX_TIMESTAMP(last_action) as last_action_unix_timestamp, prev_action, (%d - UNIX_TIMESTAMP(last_action)) AS announcetime, UNIX_TIMESTAMP(prev_action) AS prevts", time.Now().Unix())

	err := o.data.DB.Model(new(model.Peer)).Select(fields).Where("torrent = ?", tid).
		Where("user_id = ?", uid).Where("peer_id = ?", peer_id).First(peer).Error
	if err != nil {
		return nil, errors.Wrap(err, "GetPeerView")
	}

	return peer, nil

}

func (o *Peer) GetPeer(ctx context.Context, tid, uid int64, ip string) (*model.Peer, error) {
	peer := new(model.Peer)

	err := o.data.DB.Model(new(model.Peer)).Where("torrent = ?", tid).
		Where("user_id = ?", uid).Where("ip = ?", ip).First(peer).Error
	if err != nil {
		return nil, errors.Wrap(err, "GetPeer")
	}

	return peer, nil

}

func (o *Peer) GetPeerListByUser(ctx context.Context, tid, uid int64) ([]*model.Peer, error) {
	peers := make([]*model.Peer, 0)
	err := o.data.DB.Model(new(model.Peer)).Where("torrent = ?", tid).Where("user_id = ?", uid).Find(&peers).Error
	if err != nil {
		return nil, errors.Wrap(err, "GetPeerListByUser")
	}
	return peers, nil
}

func (o *Peer) CountPeersByUserAndSeedType(ctx context.Context, uid int64, seederType string) (int64, error) {
	cnt := int64(0)
	err := o.data.DB.Model(new(model.Peer)).Where("seeder = ?", seederType).Where("user_id = ?", uid).Count(&cnt).Error
	if err != nil {
		return cnt, errors.Wrap(err, "GetPeerListByUser")
	}
	return cnt, nil
}

func (o *Peer) Delete(ctx context.Context, id int64) (int64, error) {
	result := o.data.DB.Model(new(model.Peer)).Where("id = ?", id).Delete(new(model.Peer))
	if result.Error != nil {
		return 0, errors.Wrap(result.Error, "Delete")
	}
	return result.RowsAffected, nil
}

func (o *Peer) xx() {

}
