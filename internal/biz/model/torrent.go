package model

import (
	"pt/internal/biz/constant"
	"time"
)

type Torrent struct {
	ID                int64      `json:"id,omitempty" bson:"id" form:"id"`
	InfoHash          string     `json:"info_hash,omitempty" bson:"info_hash" form:"info_hash"`
	Name              string     `json:"name,omitempty" bson:"name" form:"name"`
	Filename          string     `json:"filename,omitempty" bson:"filename" form:"filename"`
	SaveAs            string     `json:"save_as,omitempty" bson:"save_as" form:"save_as"`
	Cover             string     `json:"cover,omitempty" bson:"cover" form:"cover"`
	Descr             string     `json:"descr,omitempty" bson:"descr" form:"descr"`
	SmallDescr        string     `json:"small_descr,omitempty" bson:"small_descr" form:"small_descr"`
	OriDescr          string     `json:"ori_descr,omitempty" bson:"ori_descr" form:"ori_descr"`
	Category          int64      `json:"category,omitempty" bson:"category" form:"category"`
	Source            int64      `json:"source,omitempty" bson:"source" form:"source"`
	Medium            int64      `json:"medium,omitempty" bson:"medium" form:"medium"`
	Codec             int64      `json:"codec,omitempty" bson:"codec" form:"codec"`
	Standard          int64      `json:"standard,omitempty" bson:"standard" form:"standard"`
	Processing        int64      `json:"processing,omitempty" bson:"processing" form:"processing"`
	Team              int64      `json:"team,omitempty" bson:"team" form:"team"`
	AudioCodec        int64      `json:"audio_codec,omitempty" bson:"audio_codec" form:"audio_codec"`
	Size              int64      `json:"size,omitempty" bson:"size" form:"size"`
	Added             *time.Time `json:"added,omitempty" bson:"added" form:"added"`
	Type              string     `json:"type,omitempty" bson:"type" form:"type"`
	NumFiles          uint16     `json:"num_files,omitempty" bson:"num_files" form:"num_files"`
	Comments          int64      `json:"comments,omitempty" bson:"comments" form:"comments"`
	Views             int64      `json:"views,omitempty" bson:"views" form:"views"`
	Hits              int64      `json:"hits,omitempty" bson:"hits" form:"hits"`
	TimesCompleted    int64      `json:"times_completed,omitempty" bson:"times_completed" form:"times_completed"`
	Leechers          int64      `json:"leechers,omitempty" bson:"leechers" form:"leechers"`
	Seeders           int64      `json:"seeders,omitempty" bson:"seeders" form:"seeders"`
	LastAction        *time.Time `json:"last_action,omitempty" bson:"last_action" form:"last_action"`
	Visible           string     `json:"visible,omitempty" bson:"visible" form:"visible"`
	Banned            string     `json:"banned,omitempty" bson:"banned" form:"banned"`
	Owner             int64      `json:"owner,omitempty" bson:"owner" form:"owner"`
	Nfo               []byte     `json:"nfo,omitempty" bson:"nfo" form:"nfo"`
	SPState           int64      `json:"sp_state,omitempty" bson:"sp_state" form:"sp_state"`
	PromotionTimeType int64      `json:"promotion_time_type,omitempty" bson:"promotion_time_type" form:"promotion_time_type"`
	PromotionUntil    *time.Time `json:"promotion_until,omitempty" bson:"promotion_until" form:"promotion_until"`
	Anonymous         string     `json:"anonymous,omitempty" bson:"anonymous" form:"anonymous"`
	URL               *int64     `json:"url,omitempty" bson:"url" form:"url"`
	PosState          string     `json:"pos_state,omitempty" bson:"pos_state" form:"pos_state"`
	PosStateUntil     *time.Time `json:"pos_state_until,omitempty" bson:"pos_state_until" form:"pos_state_until"`
	CacheStamp        int64      `json:"cache_stamp,omitempty" bson:"cache_stamp" form:"cache_stamp"`
	PickType          string     `json:"pick_type,omitempty" bson:"pick_type" form:"pick_type"`
	PickTime          *time.Time `json:"pick_time,omitempty" bson:"pick_time" form:"pick_time"`
	LastReseed        *time.Time `json:"last_reseed,omitempty" bson:"last_reseed" form:"last_reseed"`
	PTGen             string     `json:"pt_gen,omitempty" bson:"pt_gen" form:"pt_gen"`
	TechnicalInfo     string     `json:"technical_info,omitempty" bson:"technical_info" form:"technical_info"`
	HR                int64      `json:"hr,omitempty" bson:"hr" form:"hr"`
	ApprovalStatus    int64      `json:"approval_status,omitempty" bson:"approval_status" form:"approval_status"`
	Price             int64      `json:"price,omitempty" bson:"price" form:"price"`
	PiecesHash        string     `json:"pieces_hash,omitempty" bson:"pieces_hash" form:"pieces_hash"`
}

func (*Torrent) TableName() string {
	return "torrents"
}

type TorrentView struct {
	ID             int64                    `gorm:"column:id"`
	Size           int64                    `gorm:"column:size"`
	Owner          int64                    `gorm:"column:owner"`
	SPState        int64                    `gorm:"column:sp_state"` // 优惠状态
	Seeders        int64                    `gorm:"column:seeders"`
	Leechers       int64                    `gorm:"column:leechers"`
	Timestamp      int64                    `gorm:"column:ts"`
	Added          time.Time                `gorm:"column:added"`
	Banned         string                   `gorm:"column:banned"`
	HR             int64                    `gorm:"column:hr"`
	ApprovalStatus constant.APPROVAL_STATUS `gorm:"column:approval_status"`
	Price          float64                  `gorm:"column:price"`
	Mode           string                   `gorm:"column:mode"`
}

type TorrentBuyLog struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UID       int64     `gorm:"column:uid;not null"`
	TorrentID int64     `gorm:"column:torrent_id;not null"`
	Price     int64     `gorm:"column:price;not null"`
	Channel   string    `gorm:"column:channel;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (*TorrentBuyLog) TableName() string {
	return "torrent_buy_logs"
}
