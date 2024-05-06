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

func (o *Torrent) Get(ctx context.Context, id int64) (*model.Torrent, error) {
	torrent := new(model.Torrent)
	err := o.data.DB.WithContext(ctx).Where("id = ?", id).Find(torrent).Error

	if err != nil {
		return torrent, errors.Wrap(err, "Get")
	}
	return torrent, nil
}

func (o *Torrent) GetBuyLogs(ctx context.Context, torrentId int64) ([]*model.TorrentBuyLog, error) {
	buyLogList := make([]*model.TorrentBuyLog, 0)
	var start = 0

	for {
		buyLogListSub := make([]*model.TorrentBuyLog, 0)
		err := o.data.DB.Model(new(model.TorrentBuyLog)).Where("torrent_id = ?", torrentId).Where("id > ?", start).Order("id").Limit(100).Find(&buyLogListSub).Error
		if err != nil {
			return nil, errors.Wrap(err, "GetBuyLogs")
		}
		if len(buyLogListSub) == 0 {
			break
		}
		buyLogList = append(buyLogList, buyLogListSub...)
		start = int(buyLogListSub[len(buyLogListSub)-1].ID)
	}

	return buyLogList, nil
}

func (o *Torrent) CreateBuyLog(ctx context.Context, log *model.TorrentBuyLog) error {
	return o.data.DB.WithContext(ctx).Create(log).Error
}
