package data

import (
	"pt/internal/biz/inter"
	"pt/internal/conf"
	comgorm "pt/internal/utils/gorm"
	comredis "pt/internal/utils/redis"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewUser,
	NewCache,
	NewTorrent,
	NewPeer,
	NewAgentAllowed,
	NewHitRuns,
	NewSnatched,
	wire.Bind(new(inter.TrackerScrapeRepo), new(*User)),
	wire.Bind(new(inter.UserRepo), new(*User)),
	wire.Bind(new(inter.CacheRepo), new(*Cache)),
	wire.Bind(new(inter.TorrentRepo), new(*Torrent)),
	wire.Bind(new(inter.PeerRepo), new(*Peer)),
	wire.Bind(new(inter.AgentAllowedRepo), new(*AgentAllowed)),
	wire.Bind(new(inter.HitRunsRepo), new(*HitRuns)),
	wire.Bind(new(inter.SnatchedRepo), new(*Snatched)),
)

// Data .
type Data struct {
	redisCli *redis.Client
	DB       *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	logger.Log(log.LevelInfo, "c", c)

	db := comgorm.NewMySQLWithPanic(c.GetMysql(), logger)
	redisCli := comredis.NewWithPanic(c.GetRedis(), logger)

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		redisCli.Close()
	}

	return &Data{
		DB:       db,
		redisCli: redisCli,
	}, cleanup, nil
}
