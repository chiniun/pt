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

func (o *Torrent) FindByHash(ctx context.Context, infoHash string) (views *model.TorrentView, err error) {
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
