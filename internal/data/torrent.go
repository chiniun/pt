package data

import (
	"context"
	"pt/internal/biz/model"

	//"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
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
	err = o.data.DB.Model(new(model.Torrent)).Raw(sql, infoHash).Scan(&views).Error
	if err != nil {
		return nil, errors.Wrap(err, "FindByHash")
	}

	return
}

func (o *Torrent) UpdateByMap(ctx context.Context, id int64, info map[string]interface{}) error {

	// $updateset[] = "times_completed = times_completed + 1";
	info["time_completed"] = gorm.Expr("time_completed + ?", info["times_completed_flag"])
	err := o.data.DB.Model(new(model.Torrent)).Where("id = ?", id).UpdateColumns(info).Error
	if err != nil {
		return errors.Wrap(err, "UpdateByMap")
	}

	return nil
}
