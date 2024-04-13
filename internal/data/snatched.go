package data

import (
	"context"
	"fmt"
	"pt/internal/biz/model"
	"time"

	//"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Snatched struct {
	data *Data
	log  *log.Helper
}

func NewSnatched(data *Data, logger log.Logger) *Snatched {
	return &Snatched{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (o *Snatched) GetSnatched(ctx context.Context, tid, uid int64) (*model.Snatched, error) {

	Snatcheds := make([]*model.Snatched, 0)

	err := o.data.DB.Model(&model.Snatched{}).Where("torrentid = ? and userid = ?", tid, uid).Order("id desc").Limit(1).Find(&Snatcheds).Error
	if err != nil {
		return nil, errors.Wrap(err, "GetSnatchedList")
	}

	if len(Snatcheds) != 1 {
		return nil, errors.New(gorm.ErrRecordNotFound.Error())
	}

	return Snatcheds[0], nil
}

func (o *Snatched) UpdateSnatchedInfo(ctx context.Context, snatchid, upload, download, left int64) error {

	//todo 这里少annouceTime
	query := `UPDATE snatched SET uploaded = uploaded + ?, downloaded = downloaded + ?, to_go = ?, last_action = ? WHERE id = ?`
	dt := time.Now()
	UpdateSQL := fmt.Sprintf(query, upload, download, left, dt, snatchid)

	err := o.data.DB.Model(&model.Snatched{}).Raw(UpdateSQL).Error
	if err != nil {
		return errors.Wrap(err, "updateSnatchedInfo")
	}

	return nil
}

func (o *Snatched) Insert(ctx context.Context, snatch *model.Snatched) error {

	return o.data.DB.Model(&model.Snatched{}).Create(snatch).Error
}

func (o *Snatched) UpdateWithMap(ctx context.Context, id int64, infoMap map[string]interface{}) error {

	return o.data.DB.Model(&model.Snatched{}).Where("id = ?", id).UpdateColumns(infoMap).Error
}
