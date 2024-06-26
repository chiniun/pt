package inter

import (
	"context"
	"pt/internal/biz/model"
)

type UserRepo interface {
	GetByPasskey(ctx context.Context, passkey string) (*model.User, error)
	GetByAuthkey(ctx context.Context, authkey string) (*model.User, error)
	Get(ctx context.Context, id int64) (*model.User, error)
	Update(ctx context.Context, user *model.User) error

	// todo 是否应该放这里
	CreateBonusLog(ctx context.Context, log *model.BonusLog) error
}
