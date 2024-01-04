package biz

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/xgfone/go-bt/tracker"
	"github.com/xgfone/go-bt"
)

type AnnounceRequest struct {
	InfoHash      string `query:"info_hash"`
	PeerID        string `query:"peer_id"`
	IP            string `query:"ip"`
	Port          uint16 `query:"port"`
	Uploaded      uint   `query:"uploaded"`
	Downloaded    uint   `query:"downloaded"`
	Left          uint   `query:"left"`
	Numwant       uint   `query:"numwant"` //TODO num want, num_want
	Key           string `query:"key"`
	Compact       bool   `query:"compact"`
	SupportCrypto bool   `query:"supportcrypto"`
	Event         string `query:"event"`
	Passkey       string `json:"passkey"`
	Authkey       string `json:"authkey"`
}

type AnnounceRdequest struct {
	tracker.AnnounceRequest
	Passkey string `json:"passkey"`
	Authkey string `json:"authkey"`
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
	GetByAuthkey(ctx context.Context, authkey string) (*User, error)

}

// TrackerUsecase is a Tracker usecase.
type TrackerAnnounceUsecase struct {
	repo  TrackerAnnounceRepo
	cache CacheRepo
	log   *log.Helper
}

// NewTrackerAnnounceUsecase new a Tracker usecase.
func NewTrackerAnnounceUsecase(
	repo TrackerAnnounceRepo,
	cache CacheRepo,
	logger log.Logger) *TrackerAnnounceUsecase {
	return &TrackerAnnounceUsecase{
		repo:  repo,
		cache: cache,
		log:   log.NewHelper(logger),
	}
}

const (
	CacheKey_TorrentNotExistsKey       = "torrent_not_exists"
	CacheKey_AuthKeyInvalidKey         = "authkey_invalid"
	CacheKey_PasskeyInvalidKey         = "passkey_invalid"
	CacheKey_IsReAnnounceKey           = "isReAnnounce"
	CacheKey_ReAnnounceCheckByAuthKey  = "reAnnounceCheckByAuthKey"
	CacheKey_ReAnnounceCheckByInfoHash = "reAnnounceCheckByInfoHash"
)

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
		userAuthenticateKey = parts[1]
		subAuthkey := fmt.Sprintf("%s|%s", authKeyTid, authKeyUid)

		// check ReAnnounce
		var isReAnnounce bool
		lockParams := map[string]string{
			"user":      authKeyUid,
			"info_hash": infoHash,
		}
		lockString := buildQueryString(lockParams)
		exist, err := o.cache.CacheLock(ctx, CacheKey_IsReAnnounceKey, lockString, 20)
		if err != nil {
			return err
		}
		if !exist { //false 键已存在
			isReAnnounce = true
		}

		if !isReAnnounce {
			exist, err = o.cache.CacheLock(ctx, CacheKey_ReAnnounceCheckByAuthKey, subAuthkey, 60)
			if err != nil {
				return err
			}
			if !exist { //false 键已存在 //TODO:
				//msg := "Request too frequent(a)"

			}
		}

		res, err := o.cache.Get(ctx, CacheKey_AuthKeyInvalidKey, authkey)
		if err != nil {
			return err // 不存在会报错吗
		}
		if len(res) != 0 { //TODO
			//msg := "Invalid authkey"
		}
	} else if passkey != "" {
		userAuthenticateKey = passkey

		res, err := o.cache.Get(ctx, CacheKey_PasskeyInvalidKey, passkey)
		if err != nil {
			return err // 不存在会报错吗
		}
		if len(res) != 0 { //TODO
			//msg := "Passkey Invalid"
		}

		lockParams := map[string]string{
			"info_hash": infoHash,
			"passkey":   passkey,
		}
		lockString := buildQueryString(lockParams)
		exist, err := o.cache.CacheLock(ctx, CacheKey_IsReAnnounceKey, lockString, 20)
		if err != nil {
			return err
		}
		if !exist { //false 键已存在
			isReAnnounce = true
		}

	} else {
		// todo
		//warn("Require passkey or authkey")
	}

	exist, err := o.cache.Get(ctx, CacheKey_TorrentNotExistsKey, infoHash)
	if err != nil { //TODO ?
		return err
	}
	if len(exist) == 0 { //false 键已存在
		//msg := "Torrent not exists"
	}

	if !isReAnnounce {
		torrentReAnnounceKey := fmt.Sprintf("reAnnounceCheckByInfoHash:%s", userAuthenticateKey)

		exist, err := o.cache.CacheLock(ctx, torrentReAnnounceKey, infoHash, 60)
		if err != nil {
			return err
		}
		if !exist { //false 键已存在 //TODO:
			//$msg = "Request too frequent(h)";
			//do_log(sprintf("[ANNOUNCE] %s key: %s already exists, value: %s", $msg, $torrentReAnnounceKey, TIMENOW));
		}
	}

	//dbconn_announce

	//check authkey
	user,err := o.repo.GetByAuthkey(ctx,authkey)
	if err != nil {
		_,errLock := o.cache.CacheLock(ctx,CacheKey_AuthKeyInvalidKey,authkey, 24 * 3600 )
		if errLock != nil {
			return errLock
		}
		return err 
	}
	passkey = user.Passkey  //todo 这里会有全局passkey赋值

	// GetIP, check port
	go-bt
	

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
		user, err := o.repo.GetByPasskey(ctx, in.Passkey)
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
