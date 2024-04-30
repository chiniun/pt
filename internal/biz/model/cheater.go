package model

import "time"

type Cheaters struct {
	ID         int64     `db:"id" json:"id,omitempty" form:"id"`
	Added      time.Time `db:"added" json:"added,omitempty" form:"added"`
	UserID     int64     `db:"userid" json:"userid,omitempty" form:"userid"`
	TorrentID  int64     `db:"torrentid" json:"torrentid,omitempty" form:"torrentid"`
	Uploaded   int64     `db:"uploaded" json:"uploaded,omitempty" form:"uploaded"`
	Downloaded int64     `db:"downloaded" json:"downloaded,omitempty" form:"downloaded"`
	Anctime    int64     `db:"anctime" json:"anctime,omitempty" form:"anctime"`
	Seeders    int64     `db:"seeders" json:"seeders,omitempty" form:"seeders"`
	Leechers   int64     `db:"leechers" json:"leechers,omitempty" form:"leechers"`
	Hit        int64     `db:"hit" json:"hit,omitempty" form:"hit"`
	Dealtby    int64     `db:"dealtby" json:"dealtby,omitempty" form:"dealtby"`
	Dealtwith  int8      `db:"dealtwith" json:"dealtwith,omitempty" form:"dealtwith"`
	Comment    string    `db:"comment" json:"comment,omitempty" form:"comment"`
}

func (o *Cheaters) TableName() string {
	return "cheaters"
}
