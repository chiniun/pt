package inter

import (
	"pt/internal/biz/model"
)

type TorrentRepo interface {
	FindByHash(infoHash string) (views *model.TorrentView, err error)
}
