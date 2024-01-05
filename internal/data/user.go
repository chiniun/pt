package data

import (
	"context"
	"pt/internal/biz"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
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

func (o *User) Create(ctx context.Context, user *biz.User) error {
	return o.data.DB.WithContext(ctx).Create(user).Error

}

func (o *User) UpdateDemo(ctx context.Context, user *biz.User) error {
	return o.data.DB.Model(&biz.User{}).WithContext(ctx).
		Where("id = ?", user.Id).
		Updates(map[string]interface{}{
			"id":      user.Id,
			"passkey": user.Passkey,
		}).Error
}

func (o *User) GetByPasskey(ctx context.Context, passkey string) (*biz.User, error) {
	var user biz.User
	err := o.data.DB.WithContext(ctx).Model(&biz.User{}).Where("passkey = ?", passkey).First(&user).Error
	if err != nil {
		return nil, errors.New(500, "dbErr", err.Error())
	}
	return &user, err
}


func (o *User) GetByAuthkey(ctx context.Context, key string) (*biz.User, error) {
	var user biz.User
	err := o.data.DB.WithContext(ctx).Model(&biz.User{}).Where("authkey = ?", key).First(&user).Error
	if err != nil {
		return nil, errors.New(500, "dbErr", err.Error())
	}
	return &user, err
}

// hard delete
func (o *User) Delete(ctx context.Context, id int64) error {
	return o.data.DB.WithContext(ctx).Delete("where id = ?", id).Error
}

// hard delete
func (o *User) HardDelete(ctx context.Context, id int64) error {
	return o.data.DB.WithContext(ctx).Unscoped().Delete("where id = ?", id).Error
}
