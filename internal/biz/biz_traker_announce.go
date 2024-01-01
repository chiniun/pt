package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type AnnounceRequest struct {
	InfoHash      string `query:"info_hash"`
	PeerID        string `query:"peer_id"`
	IP            string `query:"ip"`
	Port          uint16 `query:"port"`
	Uploaded      uint   `query:"uploaded"`
	Downloaded    uint   `query:"downloaded"`
	Left          uint   `query:"left"`
	Numwant       uint   `query:"numwant"`
	Key           string `query:"key"`
	Compact       bool   `query:"compact"`
	SupportCrypto bool   `query:"supportcrypto"`
	Event         string `query:"event"`
}

func (req *AnnounceRequest) IsSeeding() bool {
	return req.Left == 0
}

type AnnounceResponse struct {
	Interval   int    `bencode:"interval"`
	Complete   int    `bencode:"complete"`
	Incomplete int    `bencode:"incomplete"`
	Peers      []byte `bencode:"peers"`
	PeersIPv6  []byte `bencode:"peers_ipv6"`
}

// TrackerAnnounceRepo 
type TrackerAnnounceRepo interface {
	//GetByPasshash(ctx context.Context, passhash string) (*User, error) 
}

// TrackerUsecase is a Tracker usecase.
type TrackerAnnounceUsecase struct {
	repo TrackerAnnounceRepo
	log  *log.Helper
}

// NewTrackerAnnounceUsecase new a Tracker usecase.
func NewTrackerAnnounceUsecase(repo TrackerAnnounceRepo, logger log.Logger) *TrackerAnnounceUsecase {
	return &TrackerAnnounceUsecase{repo: repo, log: log.NewHelper(logger)}
}

// AnounceCheck, data check before response
func (uc *TrackerAnnounceUsecase) AnounceCheck(ctx context.Context, in *AnnounceRequest) error {

	//TODO: 1 authKey || passKey check
	//TODO: 2 check exist "infoHash" which is hash of one torrent
	//3 store in redis

	//4 db passkey is equal to requst passkey?

	// GetIP, check port

	// Disable compact announce with IPv6

	// 	$peerIPV46 = "";
	// if ($ipv4) {
	//     $peerIPV46 .= ", ipv4 = " . sqlesc($ipv4);
	// }
	// if ($ipv6) {
	//     $peerIPV46 .= ", ipv6 = " . sqlesc($ipv6);
	// }

	// check port and connectable
	// if (portblacklisted($port))
	// warn("Port $port is blacklisted.");

	return nil
}

func (uc *TrackerAnnounceUsecase) AnounceHandler(ctx context.Context, in *AnnounceRequest) (*AnnounceResponse, error) {
	return nil, nil
}
