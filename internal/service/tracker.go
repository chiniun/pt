package service

import (
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

func (s *TrackerService) AppendToServer(srv *http.Server) {
	router := srv.Route("/pt/")

	// 授权
	router.GET("/announce", s.Announce) // 步骤1中台首次安装url,后端验证完跳至前端authoration界面,用户填完domain后请求步骤2接口
	router.GET("/scrape", s.Scrape)     // 步骤5, 中台confirm后回调应用的接口,应用获取中台信息

}

// Sends a greeting
func (o *TrackerService) Announce(ctx http.Context) error {

	_, err := o.auc.AnounceHandler(ctx)
	if err != nil {
		log.Errorf("%#+v", err)
		return err
	}

	return nil
}

func (o *TrackerService) Scrape(ctx http.Context) error {
	panic("not implemented") // TODO: Implement
}
