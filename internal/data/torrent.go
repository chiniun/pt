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

// func (o *Torrent) Create(ctx context.Context, Torrent *model.Torrent) error {
// 	err := o.data.DB.WithContext(ctx).Create(Torrent).Error
// 	if err != nil {
// 		return errors.Wrap(err, "Create")
// 	}

// 	return nil
// }

func (o *Torrent) FindByHash(infoHash string) (views *model.TorrentView, err error) {
	sql := `
		SELECT 
			torrents.id, 
			size,
			owner,
			sp_state,
			seeders, 
			leechers, 
			UNIX_TIMESTAMP(added) AS ts, 
			added, 
			banned, 
			hr, 
			approval_status, 
			price, 
			categories.mode 
	   	FROM 
	   		torrents 
	   	LEFT JOIN 
	   		categories on torrents.category = categories.id 
	   	WHERE 
	   		info_hash = ?
	   `
	err = o.data.DB.Raw(sql, infoHash).Scan(&views).Error
	if err != nil {
		return nil, errors.Wrap(err, "FindByHash")
	}

	return
}

func (o *Torrent) GetPeerList(ctx context.Context, tid int64, onlyLeechQuery, limit string) ([]*model.PeerView, error) {

	peers := make([]*model.PeerView, 0)

	fields := fmt.Sprintf("id, seeder, peer_id, ip, ipv4, ipv6, port, uploaded, downloaded, userid, last_action, UNIX_TIMESTAMP(last_action) as last_action_unix_timestamp, prev_action, (%d - UNIX_TIMESTAMP(last_action)) AS announcetime, UNIX_TIMESTAMP(prev_action) AS prevts", time.Now().Unix())

	peerListSQL := fmt.Sprintf("SELECT %s FROM peers WHERE torrent = %d %s %s", fields, tid, onlyLeechQuery, limit)

	err := o.data.DB.Raw(peerListSQL).Scan(&peers).Error
	if err != nil {
		return nil, errors.Wrap(err, "GetPeerList")
	}
	return peers, nil

}

func (o *Torrent) GetPeer(ctx context.Context, tid, uid int64, peer_id string) (*model.PeerView, error) {
	peers := make([]*model.PeerView, 0)

	fields := fmt.Sprintf("id, seeder, peer_id, ip, ipv4, ipv6, port, uploaded, downloaded, userid, last_action, UNIX_TIMESTAMP(last_action) as last_action_unix_timestamp, prev_action, (%d - UNIX_TIMESTAMP(last_action)) AS announcetime, UNIX_TIMESTAMP(prev_action) AS prevts", time.Now().Unix())

	peerListSQL := fmt.Sprintf("SELECT %s FROM peers WHERE torrent = %d %s %d", fields, tid, peer_id, uid)

	err := o.data.DB.Raw(peerListSQL).Scan(&peers).Error
	if err != nil {
		return nil, errors.Wrap(err, "GetPeerList")
	}

	if len(peers) >= 1 {
		return peers[0], nil
	}

	return &model.PeerView{}, nil

}
