package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/xgfone/go-bt/tracker"
	"gorm.io/gorm"

	"pt/internal/biz/inter"
	"pt/internal/biz/model"
)

type AnnounceRequest struct {
	InfoHash      string `query:"info_hash" json:"info_hash,omitempty" bson:"info_hash" form:"info_hash"`
	PeerID        string `query:"peer_id" json:"peer_id,omitempty" bson:"peer_id" form:"peer_id"`
	IP            string `query:"ip" json:"ip,omitempty" bson:"ip" form:"ip"`
	Port          uint16 `query:"port" json:"port,omitempty" bson:"port" form:"port"`
	Uploaded      uint   `query:"uploaded" json:"uploaded,omitempty" bson:"uploaded" form:"uploaded"`
	Downloaded    uint   `query:"downloaded" json:"downloaded,omitempty" bson:"downloaded" form:"downloaded"`
	Left          uint   `query:"left" json:"left,omitempty" bson:"left" form:"left"`
	Numwant       uint   `query:"numwant" json:"numwant,omitempty" bson:"numwant" form:"numwant"` //TODO num want, num_want
	Key           string `query:"key" json:"key,omitempty" bson:"key" form:"key"`
	Compact       bool   `query:"compact" json:"compact,omitempty" bson:"compact" form:"compact"`
	SupportCrypto bool   `query:"supportcrypto" json:"support_crypto,omitempty" bson:"support_crypto" form:"support_crypto"`
	Event         string `query:"event" json:"event,omitempty" bson:"event" form:"event"`

	Passkey  string `json:"passkey,omitempty" bson:"passkey" form:"passkey"`
	Authkey  string `json:"authkey,omitempty" bson:"authkey" form:"authkey"`
	RawQuery string `json:"raw_query,omitempty" bson:"raw_query" form:"raw_query"`
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
// TrackerUsecase is a Tracker usecase.
type TrackerAnnounceUsecase struct {
	urepo inter.UserRepo
	trepo inter.TorrentRepo
	cache inter.CacheRepo
	log   *log.Helper
}

// NewTrackerAnnounceUsecase new a Tracker usecase.
func NewTrackerAnnounceUsecase(
	urepo inter.UserRepo,
	cache inter.CacheRepo,
	trepo inter.TorrentRepo,
	logger log.Logger) *TrackerAnnounceUsecase {
	return &TrackerAnnounceUsecase{
		urepo: urepo,
		trepo: trepo,
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

	CacheKey_UserPasskeyContent    = "user_passkey_%s_content"
	CacheKey_TorrentHashkeyContent = "torrent_hash_%s_content"
)

// AnounceCheck, data check before response
func (o *TrackerAnnounceUsecase) AnounceHandler(ctx context.Context, in *AnnounceRequest) error {

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
		exist, err := o.cache.Lock(ctx, CacheKey_IsReAnnounceKey, lockString, 20)
		if err != nil {
			return err
		}
		if !exist { //false 键已存在
			isReAnnounce = true
		}

		if !isReAnnounce {
			exist, err = o.cache.Lock(ctx, CacheKey_ReAnnounceCheckByAuthKey, subAuthkey, 60)
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
		exist, err := o.cache.Lock(ctx, CacheKey_IsReAnnounceKey, lockString, 20)
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

		exist, err := o.cache.Lock(ctx, torrentReAnnounceKey, infoHash, 60)
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
	user, err := o.urepo.GetByAuthkey(ctx, authkey)
	if err != nil {
		_, errLock := o.cache.Lock(ctx, CacheKey_AuthKeyInvalidKey, authkey, 24*3600)
		if errLock != nil {
			return errLock
		}
		return err
	}
	passkey = user.Passkey //todo 这里会有全局passkey赋值

	// GetIP, check port

	if portBlacklisted(in.Port) { //TODO
		//warn port is blacklisted
	}

	// return peer list limit
	rsize := 50

	// seeder
	var seeder = "no"
	if in.Left == 0 {
		seeder = "yes"
	}

	o.log.Info(seeder, rsize)

	uInfoCacheKey := contentKeyCombine(CacheKey_UserPasskeyContent, passkey)
	azStr, err := o.cache.GetByKey(ctx, uInfoCacheKey)
	if err != nil {
		return err
	}
	if len(azStr) == 0 {
		user, err := o.urepo.GetByPasskey(ctx, passkey)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				_, errLock := o.cache.Lock(ctx, CacheKey_PasskeyInvalidKey, passkey, 24*3600)
				if errLock != nil {
					return nil
				}
				//TODO return error
			}
			return err
		}
		ud, err := json.Marshal(user)
		if err != nil {
			return err
		}
		_, err = o.cache.Set(ctx, uInfoCacheKey, string(ud), 3600)
		if err != nil {
			return err
		}
	}

	// checkclient
	// checkUserAgent

	// checkTorrent
	toData, err := o.cache.GetByKey(ctx, 
		contentKeyCombine(CacheKey_TorrentHashkeyContent, infoHash))
	if err != nil {
		if !errors.As(err, redis.Nil) {
			return err
		}
		// nil  查询数据库
		torrent, err := o.trepo.FindByHash(infoHash)
		if err != nil { //torrent不存在
			firstNeedle := "info_hash="
			queryString := r.URL.RawQuery
			start := strings.Index(queryString, firstNeedle) + len(firstNeedle)
			end := strings.Index(queryString[start:], "&")
			if end == -1 {
				end = len(queryString)
			} else {
				end += start
			}
			infoHashUrlEncode := queryString[start:end]
			o.log.Errorf("[TORRENT NOT EXISTS] params: %s, infoHashUrlEncode: %s\n", queryString, infoHashUrlEncode)
			o.cache.Set(ctx, 
				fmt.Sprintf("%s:%s", CacheKey_TorrentNotExistsKey, infoHashUrlEncode), time.Now().Format(time.RFC3339), 24*3600)

			return errors.Wrap(err, "torrent not registered with this tracker")
		}

		tobyte, _ := json.Marshal(torrent)
		toData = string(tobyte)
		o.cache.Set(ctx, contentKeyCombine(CacheKey_TorrentHashkeyContent, infoHash), toData, 350)
	}
	torrent := new(model.TorrentView)
	err = json.Unmarshal([]byte(toData), torrent)
	if err != nil {
		return errors.Wrap(err, "Unmarshal")
	}

	return nil
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

func portBlacklisted(port uint16) bool {
	// direct connect
	if port >= 411 && port <= 413 {
		return true
	}
	// bittorrent
	if port >= 6881 && port <= 6889 {
		return true
	}
	// kazaa
	if port == 1214 {
		return true
	}
	// gnutella
	if port >= 6346 && port <= 6347 {
		return true
	}
	// emule
	if port == 4662 {
		return true
	}
	// winmx
	if port == 6699 {
		return true
	}
	return false
}

func contentKeyCombine(key, body string) string {
	return fmt.Sprintf(key, body)
}
