package service

import (
	"context"
	v1 "pt/api/pt/v1"
	"pt/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
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
	vals := ctx.(http.Context).Request().URL.RawQuery
	temp := &biz.AnnounceRequest{
		InfoHash:      in.GetInfoHash(),
		PeerID:        in.GetPeerId(),
		IP:            in.GetIp(),
		Port:          uint16(in.GetPort()),
		Uploaded:      uint(in.GetUploaded()),
		Downloaded:    uint(in.GetDownloaded()),
		Left:          uint(in.GetLeft()),
		Numwant:       uint(in.GetNumwant()),
		Key:           in.GetKey(),
		Compact:       in.GetCompact(),
		SupportCrypto: in.GetSupportcrypto(),
		Event:         in.GetEvent(),
		Passkey:       in.GetPasskey(),
		Authkey:       in.GetAuthkey(),
		RawQuery:      vals,
	}

	err := o.auc.AnounceHandler(ctx, temp)
	if err != nil {
		log.Errorf("%#+v", err)
		return nil, err
	}

	return nil, nil
}

func (o *TrackerService) Scrape(ctx context.Context, in *v1.ScrapeRequest) (*v1.ScrapeReply, error) {
	panic("not implemented") // TODO: Implement
}
