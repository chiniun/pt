package gorm

import (
	"context"
	"os"
	"time"

	"log"

	common "pt/internal/conf"

	_ "pt/internal/utils/types/db"
	kratoslog "github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/durationpb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type IMySQL interface {
	GetUser() string
	GetPassword() string
	GetNet() string
	GetAddr() string
	GetDbname() string
	GetParams() map[string]string
	GetCollation() string
	GetLoc() string
	GetMaxAllowedPacket() uint32
	GetServerPubKey() string
	GetTlsConfig() string
	GetTimeout() *durationpb.Duration
	GetReadTimeout() *durationpb.Duration
	GetWriteTimeout() *durationpb.Duration
	GetAllowOldPasswords() bool
	GetCheckConnLiveness() bool
	GetClientFoundRows() bool
	GetColumnsWithAlias() bool
	GetInterpolateParams() bool
	GetMultiStatements() bool
	GetParseTime() bool
	GetRejectReadOnly() bool
	GetAllowNativePasswords() bool
	GetMaxIdleConns() uint64
	GetMaxOpenConns() uint64
	GetConnMaxLifetime() *durationpb.Duration
	GetDebug() bool
	GetDsn() string
	GetLog() *common.Mysql_Log
}

func NewMySQLWithPanic(cfg IMySQL, logger kratoslog.Logger) (db *gorm.DB) {
	db, err := NewMySQL(cfg, logger)
	if err != nil {
		panic(err)
	}
	return
}

func NewMySQL(cfg IMySQL, logger kratoslog.Logger) (db *gorm.DB, err error) {
	var dsn string
	if cfg.GetDsn() == "" {
		dsn = new(DsnConfig).FormConfig(cfg).FormatDSN()
	} else {
		dsn = cfg.GetDsn()
	}

	var (
		gormLoggerConfig = gormlogger.Config{ // default config
			SlowThreshold:             0 * time.Millisecond,
			LogLevel:                  gormlogger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		}
		gormConfig = &gorm.Config{}
	)
	logCfg := cfg.GetLog()

	if logCfg != nil {
		if logCfg.GetShowSlowSql() {
			gormLoggerConfig.SlowThreshold = 200 * time.Millisecond
		}

		if !logCfg.GetIgnoreRecordNotFoundError() {
			gormLoggerConfig.IgnoreRecordNotFoundError = false
		}
	}

	gormConfig.Logger = gormlogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		gormLoggerConfig,
	)

	db, err = gorm.Open(mysql.Open(dsn), gormConfig)
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
		maxIdleConns    uint64
		maxOpenConns    uint64
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

	kratoslog.NewHelper(logger).Infow("message", "Connected to MySQL")
	return
}

func Close(ctx context.Context, db *gorm.DB) (err error) {
	closeErrChan := make(chan error, 1)
	go func() {
		var err error

		defer func() {
			closeErrChan <- err
		}()

		sqlDB, err := db.DB()
		if err != nil {
			return
		}
		err = sqlDB.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case closeErr := <-closeErrChan:
			return closeErr
		}
	}
}
