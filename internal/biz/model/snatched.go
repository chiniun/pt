package model

import "time"

type Snatched struct {
	ID         int64     `json:"id"`
	Torrentid  int64     `json:"torrentid"`
	UserID     int64     `json:"userid"`
	IP         string    `json:"ip"`
	Port       uint16    `json:"port"`
	Uploaded   int64     `json:"uploaded"`
	Downloaded int64     `json:"downloaded"`
	ToGo       int64     `json:"to_go"`
	SeedTime   int64     `json:"seedtime"`
	LeechTime  int64     `json:"leechtime"`
	LastAction time.Time `json:"last_action"`
	StartAt    time.Time `json:"startat"`
	CompleteAt time.Time `json:"completeat"`
	Finished   string    `json:"finished"`
}

func (*Snatched) TableName() string {
	return "snatched"
}
