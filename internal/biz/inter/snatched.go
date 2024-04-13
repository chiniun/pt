package inter

import (
	"context"
	"pt/internal/biz/model"
)

type SnatchedRepo interface {
	GetSnatched(ctx context.Context, tid, uid int64) (*model.Snatched, error)
	Insert(ctx context.Context, snatch *model.Snatched) error
	UpdateSnatchedInfo(ctx context.Context, snatchid, upload, download, left int64) error
	UpdateWithMap(ctx context.Context, id int64, infoMap map[string]interface{}) error
}
