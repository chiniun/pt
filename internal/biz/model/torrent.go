package model

import "time"

type Torrent struct {
	ID                int32      `json:"id,omitempty" bson:"id" form:"id"`
	InfoHash          string     `json:"info_hash,omitempty" bson:"info_hash" form:"info_hash"`
	Name              string     `json:"name,omitempty" bson:"name" form:"name"`
	Filename          string     `json:"filename,omitempty" bson:"filename" form:"filename"`
	SaveAs            string     `json:"save_as,omitempty" bson:"save_as" form:"save_as"`
	Cover             string     `json:"cover,omitempty" bson:"cover" form:"cover"`
	Descr             string     `json:"descr,omitempty" bson:"descr" form:"descr"`
	SmallDescr        string     `json:"small_descr,omitempty" bson:"small_descr" form:"small_descr"`
	OriDescr          string     `json:"ori_descr,omitempty" bson:"ori_descr" form:"ori_descr"`
	Category          uint16     `json:"category,omitempty" bson:"category" form:"category"`
	Source            uint8      `json:"source,omitempty" bson:"source" form:"source"`
	Medium            uint8      `json:"medium,omitempty" bson:"medium" form:"medium"`
	Codec             uint8      `json:"codec,omitempty" bson:"codec" form:"codec"`
	Standard          uint8      `json:"standard,omitempty" bson:"standard" form:"standard"`
	Processing        uint8      `json:"processing,omitempty" bson:"processing" form:"processing"`
	Team              uint8      `json:"team,omitempty" bson:"team" form:"team"`
	AudioCodec        uint8      `json:"audio_codec,omitempty" bson:"audio_codec" form:"audio_codec"`
	Size              uint64     `json:"size,omitempty" bson:"size" form:"size"`
	Added             *time.Time `json:"added,omitempty" bson:"added" form:"added"`
	Type              string     `json:"type,omitempty" bson:"type" form:"type"`
	NumFiles          uint16     `json:"num_files,omitempty" bson:"num_files" form:"num_files"`
	Comments          int32      `json:"comments,omitempty" bson:"comments" form:"comments"`
	Views             uint32     `json:"views,omitempty" bson:"views" form:"views"`
	Hits              uint32     `json:"hits,omitempty" bson:"hits" form:"hits"`
	TimesCompleted    int32      `json:"times_completed,omitempty" bson:"times_completed" form:"times_completed"`
	Leechers          int32      `json:"leechers,omitempty" bson:"leechers" form:"leechers"`
	Seeders           int32      `json:"seeders,omitempty" bson:"seeders" form:"seeders"`
	LastAction        *time.Time `json:"last_action,omitempty" bson:"last_action" form:"last_action"`
	Visible           string     `json:"visible,omitempty" bson:"visible" form:"visible"`
	Banned            string     `json:"banned,omitempty" bson:"banned" form:"banned"`
	Owner             int32      `json:"owner,omitempty" bson:"owner" form:"owner"`
	Nfo               []byte     `json:"nfo,omitempty" bson:"nfo" form:"nfo"`
	SPState           uint8      `json:"sp_state,omitempty" bson:"sp_state" form:"sp_state"`
	PromotionTimeType uint8      `json:"promotion_time_type,omitempty" bson:"promotion_time_type" form:"promotion_time_type"`
	PromotionUntil    *time.Time `json:"promotion_until,omitempty" bson:"promotion_until" form:"promotion_until"`
	Anonymous         string     `json:"anonymous,omitempty" bson:"anonymous" form:"anonymous"`
	URL               *uint32    `json:"url,omitempty" bson:"url" form:"url"`
	PosState          string     `json:"pos_state,omitempty" bson:"pos_state" form:"pos_state"`
	PosStateUntil     *time.Time `json:"pos_state_until,omitempty" bson:"pos_state_until" form:"pos_state_until"`
	CacheStamp        uint8      `json:"cache_stamp,omitempty" bson:"cache_stamp" form:"cache_stamp"`
	PickType          string     `json:"pick_type,omitempty" bson:"pick_type" form:"pick_type"`
	PickTime          *time.Time `json:"pick_time,omitempty" bson:"pick_time" form:"pick_time"`
	LastReseed        *time.Time `json:"last_reseed,omitempty" bson:"last_reseed" form:"last_reseed"`
	PTGen             string     `json:"pt_gen,omitempty" bson:"pt_gen" form:"pt_gen"`
	TechnicalInfo     string     `json:"technical_info,omitempty" bson:"technical_info" form:"technical_info"`
	HR                int8       `json:"hr,omitempty" bson:"hr" form:"hr"`
	ApprovalStatus    int32      `json:"approval_status,omitempty" bson:"approval_status" form:"approval_status"`
	Price             int32      `json:"price,omitempty" bson:"price" form:"price"`
	PiecesHash        string     `json:"pieces_hash,omitempty" bson:"pieces_hash" form:"pieces_hash"`
}

func (*Torrent) TableName() string {
	return "torrents"
}

type TorrentView struct {
	ID             int32     `gorm:"column:id"`
	Size           uint64    `gorm:"column:size"`
	Owner          int32     `gorm:"column:owner"`
	SPState        uint8     `gorm:"column:sp_state"`
	Seeders        int32     `gorm:"column:seeders"`
	Leechers       int32     `gorm:"column:leechers"`
	Timestamp      int64     `gorm:"column:ts"`
	Added          time.Time `gorm:"column:added"`
	Banned         string    `gorm:"column:banned"`
	HR             int8      `gorm:"column:hr"`
	ApprovalStatus int32     `gorm:"column:approval_status"`
	Price          int32     `gorm:"column:price"`
	CategoryMode   string    `gorm:"column:mode"`
}