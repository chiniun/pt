package service

import (
	"context"
	v1 "pt/api/pt/v1"
	"pt/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// TrackerService is a Auth service.
type TrackerService struct {
	v1.UnimplementedTrackerServer
	auc *biz.TrackerAnnounceUsecase
	suc *biz.TrackerScrapeUsecase
	log *log.Helper
}

// NewTrackerService new a Auth service.
func NewTrackerService(
	auc *biz.TrackerAnnounceUsecase,
	suc *biz.TrackerScrapeUsecase,
	logger log.Logger,
) *TrackerService {
	return &TrackerService{
		auc: auc,
		suc: suc,
		log: log.NewHelper(logger),
	}
}

// Sends a greeting
func (o *TrackerService) Announce(ctx context.Context, in *v1.AnnounceRequest) (*v1.AnnounceReply, error) {
	panic("not implemented") // TODO: Implement
}

func (o *TrackerService) Scrape(ctx context.Context, in *v1.ScrapeRequest) (*v1.ScrapeReply, error) {
	panic("not implemented") // TODO: Implement
}
