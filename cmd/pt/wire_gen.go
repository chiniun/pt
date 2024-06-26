// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"pt/internal/biz"
	"pt/internal/conf"
	"pt/internal/data"
	"pt/internal/server"
	"pt/internal/service"
)

import (
	_ "go.uber.org/automaxprocs"
)

// Injectors from wire.go:

// wireApp init kratos application.
func wireApp(confServer *conf.Server, confData *conf.Data, logger log.Logger) (*kratos.App, func(), error) {
	greeterService := service.NewGreeterService()
	dataData, cleanup, err := data.NewData(confData, logger)
	if err != nil {
		return nil, nil, err
	}
	user := data.NewUser(dataData, logger)
	cache := data.NewCache(dataData, logger)
	peer := data.NewPeer(dataData, logger)
	torrent := data.NewTorrent(dataData, logger)
	snatched := data.NewSnatched(dataData, logger)
	hitRuns := data.NewHitRuns(dataData, logger)
	agentAllowed := data.NewAgentAllowed(dataData, logger)
	agentDeny := data.NewAgentDeny(dataData, logger)
	cheaters := data.NewCheaters(dataData, logger)
	seedBox := data.NewSeedBox(dataData, logger)
	trackerAnnounceUsecase := biz.NewTrackerAnnounceUsecase(user, cache, peer, torrent, snatched, hitRuns, agentAllowed, agentDeny, cheaters, seedBox, logger)
	trackerScrapeUsecase := biz.NewTrackerScrapeUsecase(user, logger)
	trackerService := service.NewTrackerService(trackerAnnounceUsecase, trackerScrapeUsecase, logger)
	grpcServer := server.NewGRPCServer(confServer, greeterService, trackerService, logger)
	httpServer := server.NewHTTPServer(confServer, greeterService, trackerService, trackerService, logger)
	app := newApp(logger, grpcServer, httpServer)
	return app, func() {
		cleanup()
	}, nil
}
