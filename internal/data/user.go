package data

import (
	"context"
	"pt/internal/biz/model"

	//"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
)

type User struct {
	data *Data
	log  *log.Helper
}

func NewUser(data *Data, logger log.Logger) *User {
	return &User{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (o *User) Create(ctx context.Context, user *model.User) error {
	return o.data.DB.WithContext(ctx).Create(user).Error

}

func (o *User) Update(ctx context.Context, user *model.User) error {
	err := o.data.DB.WithContext(ctx).Updates(user).Error
	if err != nil {
		return errors.Wrap(err, "update")
	}
	return nil
}

func (o *User) Get(ctx context.Context, id int64) (*model.User, error) {
	user := new(model.User)
	return user, o.data.DB.WithContext(ctx).Where("id = ?", id).Find(user).Error
}

func (o *User) UpdateDemo(ctx context.Context, user *model.User) error {
	return o.data.DB.Model(&model.User{}).WithContext(ctx).
		Where("id = ?", user.Id).
		Updates(map[string]interface{}{
			"id":      user.Id,
			"passkey": user.Passkey,
		}).Error
}

func (o *User) GetByPasskey(ctx context.Context, passkey string) (*model.User, error) {
	var user model.User
	err := o.data.DB.WithContext(ctx).Model(&model.User{}).Where("passkey = ?", passkey).First(&user).Error
	if err != nil {
		return nil, errors.Wrap(err, "GetByPasskey")
	}
	return &user, err
}

func (o *User) GetByAuthkey(ctx context.Context, key string) (*model.User, error) {
	var user model.User
	err := o.data.DB.WithContext(ctx).Model(&model.User{}).Where("authkey = ?", key).First(&user).Error
	if err != nil {
		return nil, errors.Wrap(err, "GetByAuthkey")
	}
	return &user, err
}

// hard delete
func (o *User) Delete(ctx context.Context, id int64) error {
	return o.data.DB.WithContext(ctx).Where("id = ?", id).Delete(&model.User{}).Error
}

// hard delete
func (o *User) HardDelete(ctx context.Context, id int64) error {
	return o.data.DB.WithContext(ctx).Unscoped().Where("id = ?", id).Delete(&model.User{}).Error
}

func (o *User) CreateBonusLog(ctx context.Context, log *model.BonusLog) error {
	return o.data.DB.WithContext(ctx).Model(new(model.BonusLog)).Create(log).Error
}
