package model

import "time"

type SeedBoxRecord struct {
	ID             int64     `db:"id"`
	Type           int64     `db:"type"`
	UID            int64     `db:"uid"`
	Status         int64     `db:"status"`
	Operator       string    `db:"operator"`
	Bandwidth      int64     `db:"bandwidth"`
	IP             string    `db:"ip"`
	IPBegin        string    `db:"ip_begin"`
	IPEnd          string    `db:"ip_end"`
	IPBeginNumeric string    `db:"ip_begin_numeric"`
	IPEndNumeric   string    `db:"ip_end_numeric"`
	Version        int64     `db:"version"`
	Comment        string    `db:"comment"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
	IsAllowed      int64     `db:"is_allowed"`
}

func (o *SeedBoxRecord) TableName() string {
	return "seed_box_records"
}
