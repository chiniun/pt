package data

import (
	"context"
	"pt/internal/biz/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
)

type seedBox struct {
	data *Data
	log  *log.Helper
}

func NewSeedBox(data *Data, logger log.Logger) *seedBox {
	return &seedBox{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (o *seedBox) Create(ctx context.Context, seedBox *model.SeedBoxRecord) error {
	err := o.data.DB.WithContext(ctx).Create(seedBox).Error
	if err != nil {
		return errors.Wrap(err, "create")
	}
	return nil
}

func (o *seedBox) Update(ctx context.Context, seedBox *model.SeedBoxRecord) error {
	err := o.data.DB.WithContext(ctx).Updates(seedBox).Error
	if err != nil {
		return errors.Wrap(err, "update")
	}
	return nil
}

func (o *seedBox) Get(ctx context.Context, id int64) (*model.SeedBoxRecord, error) {
	seedBox := new(model.SeedBoxRecord)
	err := o.data.DB.WithContext(ctx).Where("id = ?", id).Find(seedBox).Error
	if err != nil {
		return nil, errors.Wrap(err, "create")
	}
	return seedBox, nil
}

func (o *seedBox)Query(ctx context.Context,seedBox *model.SeedBoxRecord)(*model.SeedBoxRecord,error)  {
	//'select id from seed_box_records where `ip_begin_numeric` <= "%s" and `ip_end_numeric` >= "%s" and `uid` = %s and `type` = %s and `version` = %s and `status` = %s and `is_allowed` = 0  limit 1',
	session := o.data.DB.WithContext(ctx).Where("ip_begin_numeric <= ?",seedBox.IPBeginNumeric).Where("ip_end_numeric >= ?",seedBox.IPEndNumeric).
	Where("version = ?",seedBox.Version).Where("status = ?",seedBox.Status).Where("is_allowed = ?",seedBox.IsAllowed)
	if seedBox.Type != 0 {
		session = session.Where("type = ?",seedBox.Type)
	}
	if seedBox.UID != 0 {
		session = session.Where("user_id = ?",seedBox.UID)
	}

	result := new(model.SeedBoxRecord)

	err := session.First(result).Error
	if err != nil {
		return nil, errors.Wrap(err, "Query")
	}
	return result, nil	

}
