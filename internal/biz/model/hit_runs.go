package model

import "time"

// hr状态
const (
	STATUS_INSPECTING = 1
	STATUS_REACHED    = 2
	STATUS_UNREACHED  = 3
	STATUS_PARDONED   = 4
)

type HitRuns struct {
	ID         int64     `json:"id"`
	UID        int64     `json:"uid"`
	TorrentID  int64     `json:"torrent_id"`
	SnatchedID int64     `json:"snatched_id"`
	Status     int       `json:"status"`
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (o *HitRuns) TableName() string {
	return "hit_and_runs"
}
