package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"pt/internal/biz/inter"
)

type ScrapeRequest struct {
	InfoHash string `query:"info_hash"`
}

type ScrapeResponse struct {
	Files map[string]Stat `bencode:"files"`
}

type Stat struct {
	Complete   int `bencode:"complete"`
	Incomplete int `bencode:"incomplete"`
	// Downloaded uint `bencode:"downloaded"`
}


// TrackerScrapeRepo is a Greater repo.
// TrackerUsecase is a Tracker usecase.
type TrackerScrapeUsecase struct {
	repo inter.TrackerScrapeRepo
	log  *log.Helper
}

// NewTrackerScrapeUsecase new a Tracker usecase.
func NewTrackerScrapeUsecase(repo inter.TrackerScrapeRepo, logger log.Logger) *TrackerScrapeUsecase {
	return &TrackerScrapeUsecase{repo: repo, log: log.NewHelper(logger)}
}

// AnounceCheck, data check before response
func (uc *TrackerScrapeUsecase) Scrape(ctx context.Context, in *ScrapeRequest) error {


	return nil
}
