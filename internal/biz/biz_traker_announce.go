package biz

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	bt "github.com/xgfone/go-bt/tracker"
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
  Passkey   string `json:"passkey"`
  Authkey   string `json:"authkey"`
 
}

type AnnounceRdequest struct{
  bt.AnnounceRequest
  Passkey   string `json:"passkey"`
  Authkey   string `json:"authkey"`
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
	GetByPasskey(ctx context.Context, passkey string) (*User, error) 
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
func (o *TrackerAnnounceUsecase) AnounceCheck(ctx context.Context, in *AnnounceRequest) error {

  
	//TODO: 1 authKey || passKey check
  isReAnnounce := false
	userAuthenticateKey := ""

	authkey := in.Authkey
	passkey := in.Passkey
	infoHash := in.InfoHash

	if authkey != "" {
		parts := strings.Split(authkey, "|")
		if len(parts) != 3 {
			o.log.Warn("authkey format error")
		}
		authKeyTid := parts[0]
		authKeyUid := parts[1]
    userAuthenticateKey := parts[1]
		subAuthkey := fmt.Sprintf("%s|%s", authKeyTid, authKeyUid)

		// check ReAnnounce
		lockParams := map[string]string{
			"user":      authKeyUid,
			"info_hash": infoHash,
		}
    
		lockString := buildQueryString(lockParams)
    

		lockKey := "isReAnnounce:" + md5.New().Write([]byte(lockString))
		// Redis code here...

		if !isReAnnounce && false { // Redis code here...
			msg := "Request too frequent(a)"
			doLog(fmt.Sprintf("[ANNOUNCE] %s key: %s already exists, value: %s", msg, reAnnounceCheckByAuthKey, TIMENOW))
			warn(msg, 300)
		}

		if false { // Redis code here...
			msg := "Invalid authkey"
			doLog("[ANNOUNCE] " + msg)
			warn(msg)
		}
	} else if passkey != "" {
		userAuthenticateKey = passkey

		if false { // Redis code here...
			msg := "Passkey invalid"
			doLog("[ANNOUNCE] " + msg)
			warn(msg)
		}

		lockParams := map[string]string{
			"info_hash": infoHash,
			"passkey":   passkey,
		}
		lockString := buildQueryString(lockParams)
		lockKey := "isReAnnounce:" + md5(lockString)
		// Redis code here...

		if !isReAnnounce && false { // Redis code here...
			msg := "Request too frequent(h)"
			doLog(fmt.Sprintf("[ANNOUNCE] %s key: %s already exists, value: %s", msg, torrentReAnnounceKey, TIMENOW))
			warn(msg, 300)
		}
	} else {
		warn("Require passkey or authkey")
	}

	if false { // Redis code here...
		msg := "Torrent not exists"
		doLog("[ANNOUNCE] " + msg)
		err(msg)
	}

	torrentReAnnounceKey := fmt.Sprintf("reAnnounceCheckByInfoHash:%s:%s", userAuthenticateKey, infoHash)
	if !isReAnnounce && false { // Redis code here...
		msg := "Request too frequent(h)"
		doLog(fmt.Sprintf("[ANNOUNCE] %s key: %s already exists, value: %s", msg, torrentReAnnounceKey, TIMENOW))
		warn(msg, 300)
	}
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
	// warn("Port $ blacklisted.");

	return nil
}

func (o *TrackerAnnounceUsecase) AnounceHandler(ctx context.Context, in *AnnounceRequest) (*AnnounceResponse, error) {

  var seeder bool
  in.AnnounceRequest.
  user,err :=o.repo.GetByPasskey(ctx,in.Passkey)
  if err != nil {
    return nil, err
  } 

  
    
  
   
	return nil, nil
}


func buildQueryString(params map[string]string) string {
	query := ""
	for key, value := range params {
		if query != "" {
			query += "&"
		}
		query += key + "=" + value
	}
	return query
}
