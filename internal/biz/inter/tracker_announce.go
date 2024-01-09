package inter

import (
	"context"
	"pt/internal/biz/model"
)

type TrackerAnnounceRepo interface {
	GetByPasskey(ctx context.Context, passkey string) (*model.User, error)
	GetByAuthkey(ctx context.Context, authkey string) (*model.User, error)
}
