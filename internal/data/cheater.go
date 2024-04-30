package data

import (
	"context"
	"pt/internal/biz/model"
	"time"

	//"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
)

type Cheaters struct {
	data *Data
	log  *log.Helper
}

func NewCheaters(data *Data, logger log.Logger) *Cheaters {
	return &Cheaters{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (o *Cheaters) Create(ctx context.Context, cheater *model.Cheaters) error {
	return o.data.DB.WithContext(ctx).Create(cheater).Error

}

func (o *Cheaters) Update(ctx context.Context, cheater *model.Cheaters) error {
	err := o.data.DB.WithContext(ctx).Updates(cheater).Error
	if err != nil {
		return errors.Wrap(err, "update")
	}
	return nil
}

func (o *Cheaters) Get(ctx context.Context, id int64) (*model.Cheaters, error) {
	cheater := new(model.Cheaters)
	return cheater, o.data.DB.WithContext(ctx).Where("id = ?", id).Find(cheater).Error
}

func (o *Cheaters) Count(ctx context.Context, uid, tid int64, added time.Time) (int64, error) {
	var cnt int64
	return cnt, o.data.DB.WithContext(ctx).Where("user_id = ?", uid).Where("tid = ?", tid).Where("added > ?", added).Count(&cnt).Error
}


func (o *Cheaters) Query(ctx context.Context, uid, tid int64, added time.Time) ([]*model.Cheaters, error) {
	cheaterList := make([]*model.Cheaters,0)
	return cheaterList, o.data.DB.WithContext(ctx).Where("user_id = ?", uid).Where("tid = ?", tid).Where("added > ?", added).Find(&cheaterList).Error
}
