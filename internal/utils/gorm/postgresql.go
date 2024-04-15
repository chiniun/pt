package gorm

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPOSTGRESQLWithPanic(cfg IMySQL, logger log.Logger) (db *gorm.DB) {
	db, err := NewPOSTGRESQL(cfg, logger)
	if err != nil {
		panic(err)
	}
	return
}
func NewPOSTGRESQL(cfg IMySQL, logger log.Logger) (db *gorm.DB, err error) {
	var dsn string
	if cfg.GetDsn() == "" {
		dsn = new(DsnConfig).FormConfig(cfg).FormatDSN()
	} else {
		dsn = cfg.GetDsn()
	}
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return
	}
	if cfg.GetDebug() {
		db = db.Debug()
	}

	sqlDB, err := db.DB()
	if err != nil {
		return
	}

	var (
		connMaxLifetime time.Duration
		maxIdleConns    int64
		maxOpenConns    int64
	)
	if !cfg.GetConnMaxLifetime().IsValid() {
		connMaxLifetime = time.Hour
	}
	if cfg.GetMaxIdleConns() == 0 {
		maxIdleConns = 10
	}
	if cfg.GetMaxOpenConns() == 0 {
		maxOpenConns = 100
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(int(maxIdleConns))
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(int(maxOpenConns))
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	log.NewHelper(logger).Infow("message", "Connected to POSTGRESQL")
	return
}
