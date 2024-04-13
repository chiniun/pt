package inter

import (
	"context"
	"pt/internal/biz/model"
)

type UserRepo interface {
	GetByPasskey(ctx context.Context, passkey string) (*model.User, error)
	GetByAuthkey(ctx context.Context, authkey string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
}
