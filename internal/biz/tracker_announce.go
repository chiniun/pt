package biz

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/xgfone/go-bt/tracker"
	"gorm.io/gorm"

	"pt/internal/biz/constant"
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
	Interval    int    `bencode:"interval"`
	MinInterval int    `bencode:"min interval"`
	Complete    int    `bencode:"complete"`
	Incomplete  int    `bencode:"incomplete"`
	Peers       []byte `bencode:"peers"`
	PeersIPv6   []byte `bencode:"peers_ipv6"`
}

// TrackerAnnounceRepo
// TrackerUsecase is a Tracker usecase.
type TrackerAnnounceUsecase struct {
	urepo inter.UserRepo
	trepo inter.TorrentRepo
	prepo inter.PeerRepo
	cache inter.CacheRepo
	log   *log.Helper
}

// NewTrackerAnnounceUsecase new a Tracker usecase.
func NewTrackerAnnounceUsecase(
	urepo inter.UserRepo,
	cache inter.CacheRepo,
	prepo inter.PeerRepo,
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
func (o *TrackerAnnounceUsecase) AnounceHandler(ctx context.Context, in *AnnounceRequest) (resp AnnounceResponse, err error) {

	//TODO: 1 authKey || passKey check

	var (
		authKeyTid          string
		authKeyUid          string
		userAuthenticateKey string
		subAuthkey          string
		isReAnnounce        bool
	)

	authkey := in.Authkey
	passkey := in.Passkey
	infoHash := in.InfoHash

	if authkey != "" {
		parts := strings.Split(authkey, "|")
		if len(parts) != 3 {
			o.log.Warn("authkey format error")
		}
		authKeyTid = parts[0]
		authKeyUid = parts[1]
		userAuthenticateKey = parts[1]
		subAuthkey = fmt.Sprintf("%s|%s", authKeyTid, authKeyUid)

		// check ReAnnounce
		var isReAnnounce bool
		lockParams := map[string]string{
			"user":      authKeyUid,
			"info_hash": infoHash,
		}
		lockString := buildQueryString(lockParams)
		exist, err := o.cache.Lock(ctx, CacheKey_IsReAnnounceKey, lockString, 20)
		if err != nil {
			return resp, err
		}
		if !exist { //false 键已存在
			isReAnnounce = true
		}

		if !isReAnnounce {
			exist, err = o.cache.Lock(ctx, CacheKey_ReAnnounceCheckByAuthKey, subAuthkey, 60)
			if err != nil {
				return resp, err
			}
			if !exist { //false 键已存在 //TODO:
				//msg := "Request too frequent(a)"

			}
		}

		res, err := o.cache.Get(ctx, CacheKey_AuthKeyInvalidKey, authkey)
		if err != nil {
			return resp, err // 不存在会报错吗
		}
		if len(res) != 0 { //TODO
			//msg := "Invalid authkey"
		}
	} else if passkey != "" {
		userAuthenticateKey = passkey

		res, err := o.cache.Get(ctx, CacheKey_PasskeyInvalidKey, passkey)
		if err != nil {
			return resp, err // 不存在会报错吗
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
			return resp, err
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
		return resp, err
	}
	if len(exist) == 0 { //false 键已存在
		//msg := "Torrent not exists"
	}

	if !isReAnnounce {
		torrentReAnnounceKey := fmt.Sprintf("reAnnounceCheckByInfoHash:%s", userAuthenticateKey)

		exist, err := o.cache.Lock(ctx, torrentReAnnounceKey, infoHash, 60)
		if err != nil {
			return resp, err
		}
		if !exist { //false 键已存在 //TODO:
			//$msg = "Request too frequent(h)";
			//do_log(sprintf("[ANNOUNCE] %s key: %s already exists, value: %s", $msg, $torrentReAnnounceKey, TIMENOW));
		}
	}

	//dbconn_announce

	//check authkey //todo authKey not exist
	user, err := o.urepo.GetByAuthkey(ctx, authkey)
	if err != nil {
		_, errLock := o.cache.Lock(ctx, CacheKey_AuthKeyInvalidKey, authkey, 24*3600)
		if errLock != nil {
			return resp, errLock
		}
		return resp, err
	}
	passkey = user.Passkey //todo 这里会有全局passkey赋值

	var compact int
	ip := in.IP
	if in.Port > 0xffff {
		return errors.New("invalid port")
	}
	p := net.ParseIP(ip) //Disable compact announce with IPv6
	if p == nil {
		compact = 0
	}

	var ipv4, ipv6 = ""
	if p != nil {
		if r := p.To4(); r != nil {
			ipv4 = ip
		} else {
			ipv6 = ip
		}
	}
	// todo 请求头里查询ipv4 或者 v6

	peerIPV46 := ""
	if ipv4 != "" {
		peerIPV46 += fmt.Sprintf(", ipv4 = %s", &ipv4)
	}
	if ipv6 != "" {
		peerIPV46 += fmt.Sprintf(", ipv6 = %s", &ipv6)
	}

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
		return resp, err
	}
	if len(azStr) == 0 {
		user, err = o.urepo.GetByPasskey(ctx, passkey)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				_, errLock := o.cache.Lock(ctx, CacheKey_PasskeyInvalidKey, passkey, 24*3600)
				if errLock != nil {
					return resp, nil
				}
				//TODO return error
			}
			return resp, err
		}
		ud, err := json.Marshal(user)
		if err != nil {
			return resp, err
		}
		_, err = o.cache.Set(ctx, uInfoCacheKey, string(ud), 3600)
		if err != nil {
			return resp, err
		}
	}

	// checkclient
	// checkUserAgent

	// checkTorrent
	toData, err := o.cache.GetByKey(ctx,
		contentKeyCombine(CacheKey_TorrentHashkeyContent, infoHash))
	if err != nil {
		if !errors.As(err, redis.Nil) {
			return resp, err
		}
		// nil  查询数据库
		torrent, err := o.trepo.FindByHash(ctx, infoHash)
		if err != nil { //torrent不存在
			firstNeedle := "info_hash="
			queryString := in.RawQuery
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

			return resp, errors.Wrap(err, "torrent not registered with this tracker")
		}

		tobyte, _ := json.Marshal(torrent)
		toData = string(tobyte)
		o.cache.Set(ctx, contentKeyCombine(CacheKey_TorrentHashkeyContent, infoHash), toData, 350)
	}
	torrent := new(model.TorrentView)
	err = json.Unmarshal([]byte(toData), torrent)
	if err != nil {
		return resp, errors.Wrap(err, "Unmarshal")
	}

	if authKeyTid != "" && authKeyTid != strconv.FormatInt(int64(torrent.ID), 10) {
		_, err = o.cache.Lock(ctx, CacheKey_AuthKeyInvalidKey, authkey, 24*3600, false)
		if err != nil {
			return resp, err
		}
		//do_log("[ANNOUNCE] $msg");
		//warn($msg);
	}

	if torrent.Banned == "yes" {
		return resp, errors.WithStack(errors.New("torrents banned"))
	}
	if torrent.ApprovalStatus != constant.APPROVAL_STATUS_ALLOW {
		return resp, errors.WithStack(errors.New("torrent review not approved"))
	}

	// select peers info from peers table for this torrent
	var (
		onlyLeechQuery string
		limit          string
	)

	numPeers := torrent.Seeders + torrent.Leechers
	var newNumPeers int64

	if seeder == "yes" {
		onlyLeechQuery = " AND seeder = 'no' "
		newNumPeers = torrent.Leechers
	} else {
		newNumPeers = numPeers
	}

	if newNumPeers > int64(rsize) {
		limit = fmt.Sprintf(" ORDER BY RAND() LIMIT %d", rsize)
	}

	var announceWait = constant.MIN_ANNOUNCE_WAIT_SECOND

	realAnnounceInterval := constant.AnnounceInterval
	if (announceWait < int(constant.AnnounceInterThree)) &&
		((time.Now().Unix() - torrent.Timestamp) >= int64(constant.AnnounceInterThreeAge*86400)) {
		realAnnounceInterval = constant.AnnounceInterThree
	} else if (announceWait < int(constant.AnnounceInterTwo)) &&
		((time.Now().Unix() - torrent.Timestamp) >= int64(constant.AnnounceInterThreeAge*86400)) {
		realAnnounceInterval = constant.AnnounceInterTwo
	}

	//rep_dict
	resp.Interval = int(realAnnounceInterval)
	resp.MinInterval = int(announceWait)
	resp.Complete = int(torrent.Seeders)
	resp.Incomplete = int(torrent.Leechers)
	resp.Peers = nil // // By default it is a array object, only when `&compact=1` then it should be a string
	resp.PeersIPv6 = nil
	// todo compact
	if compact == 1 {
		resp.Peers = []byte("")     // Change `peers` from array to string
		resp.PeersIPv6 = []byte("") // If peer use IPv6 address , we should add packed string in `peers6`
	}

	if isReAnnounce {
		o.log.Error("YES_RE_ANNOUNCE")
		return
	}

	// GetPeerList
	peers, err := o.trepo.GetPeerList(ctx, torrent.ID, onlyLeechQuery, limit)
	if err != nil {
		return
	}
	selfPeer := &model.PeerView{}
	if in.Event == "stop" {

	} else {
		for _, peer := range peers {
			peer.PeerID = hashPad(peer.PeerID)
			if peer.PeerID == in.PeerID && peer.UserID == uint32(user.Id) {
				selfPeer = peer
				continue
			}

			if compact == 1 {
				if peer.IPv4 != "" {
					resp.Peers = append(resp.Peers, concatIPAndPort(peer.IPv4, peer.Port)...)
				}
				if peer.IPv6 != "" {
					resp.PeersIPv6 = append(resp.PeersIPv6, concatIPAndPort(peer.IPv6, peer.Port)...)
				}
			} else {
				if peer.IPv4 != "" {
					tmpPeer := model.PeerBin{
						PeerID: peer.PeerID,
						IP:     peer.IPv4,
						Port:   int32(peer.Port),
					}
					resp.Peers = append(resp.Peers, tmpPeer)
				}

				if peer.IPv6 != "" {
					tmpPeer := model.PeerBin{
						PeerID: peer.PeerID,
						IP:     peer.IPv6,
						Port:   int32(peer.Port),
					}
					resp.Peers = append(resp.Peers, tmpPeer) //todo 这里也用Peers吗
				}
			}

		}

	}

	if selfPeer.ID == 0 {
		selfPeer, err = o.prepo.GetPeerView(ctx, torrent.ID, user.Id, in.PeerID)
		if err != nil {
			return resp, err
		}
	}

	if selfPeer.ID != 0 && in.Event != "" && selfPeer.Prevts > (time.Now().Unix()-int64(announceWait)) {
		//warn('There is a minimum announce time of ' . $announce_wait . ' seconds', $announce_wait);
	}

	var isSeedBoxRuleEnabled bool //TODO 配置文件里拿
	var isIPSeedBox bool          // TODO 盒子判断

	if isSeedBoxRuleEnabled {
		if ipv4 != "" {
			isIPSeedBox = false
		}
	}
	o.log.Infow("isIPSeedBox", isIPSeedBox)

	if selfPeer.ID == 0 {
		sameIPRecord, err := o.prepo.GetPeer(ctx, torrent.ID, user.Id, ip)
		if err == nil && sameIPRecord.ID != 0 && seeder == "yes" {
			return resp, errors.New("You cannot seed the same torrent in the same location from more than 1 client.")
		}

		peers, err := o.prepo.GetPeerListByUser(ctx, torrent.ID, user.Id)
		if err != nil {
			return resp, err
		}

		if len(peers) >= 1 && seeder == "no" {
			return resp, errors.New("You already are downloading the same torrent. You may only leech from one location at a time.")
		}

		if len(peers) >= 3 && seeder == "yes" {
			return resp, errors.New("You cannot seed the same torrent from more than 3 locations.")
		}

		if user.Enabled == "no" {
			return resp, errors.New("Your account is disabled!")
		} else if user.Parked == "Yes" {
			return resp, errors.New("Your account is parked! (Read the FAQ)")
		} else if user.DownloadPos == "no" {
			return resp, errors.New("Your downloading privileges have been disabled! (Read the rules)")
		}

		// 非vip权限校验
		if user.Class < constant.UC_VIP {
			var ratio float64 = 999999999999999
			if user.Downloaded > 0 {
				ratio = float64(user.Uploaded) / float64(user.Downloaded)
			}

			var gigs = user.Downloaded / (1024 * 1024 * 1024)

			var waitsystem string // yes or no //TODO GetConfig
			var maxDlsSystem string
			var wait int64

			if waitsystem == "yes" {
				if gigs > 10 {
					if ratio < 0.4 {
						wait = 24
					} else if ratio < 0.5 {
						wait = 12
					} else if ratio < 0.6 {
						wait = 6
					} else if ratio < 0.8 {
						wait = 3
					} else {
						wait = 0
					}

					var elapsed = time.Now().Unix() - torrent.Timestamp // TODO
					if elapsed < wait {
						return resp, errors.New(fmt.Sprintf("Your ratio is too low! You need to wait %d to start", wait*3600-elapsed))
					}
				}
			}
			var max int64
			if maxDlsSystem == "yes" {
				if gigs > 10 {
					if ratio < 0.5 {
						max = 1
					} else if ratio < 0.65 {
						max = 2
					} else if ratio < 0.8 {
						max = 3
					} else if ratio < 0.95 {
						max = 4
					} else {
						max = 0
					}
					if max > 0 {
						peerCnt, err := o.prepo.CountPeersByUserAndSeedType(ctx, user.Id, "no")
						if err != nil {
							return resp, err
						}
						if peerCnt >= max {
							return resp, errors.New(fmt.Sprintf("Your slot limit is reached! You may at most download %d torrents at the same time", max))
						}
					}
				}
			}
		}

		var ConfigPaidTorrentEnabled bool
		var flag = seeder == "no" &&
			user.SeedBonus != 0 &&
			torrent.Price > 0 &&
			torrent.Owner != user.Id &&
			ConfigPaidTorrentEnabled

		if flag { //redisLock
			hasBuy, err := o.cache.Get(ctx, constant.CACHE_KEY_BOUGHT_USER_PREFIX, strconv.FormatInt(torrent.ID, 10))
			if err != nil {
				if errors.As(err, redis.Nil) {

				}
				return resp, err
			}

		}
	} else {
		var upthis, trueUpthis float64
		upthis = Max(0, float64(in.Uploaded-uint(selfPeer.Uploaded)))
		trueUpthis = upthis

		var announcetime int64
		if selfPeer.Seeder == "yes" {
			announcetime = selfPeer.Announcetime
		} else {

		}

	}

	// set non-type event
	var event string
	if selfPeer.ID != 0 && in.Event == "stopped" {
		err := o.prepo.Delete(ctx, selfPeer.ID)
		if err != nil {
			o.log.Error(err)
		} else { // todo updateSnatched

		}

	} else if selfPeer.ID != 0 {

	} else {

	}

	//handle hr
	var isDonor bool
	if in.Left > 0 || event == "completed" &&
		user.Class < constant.UC_VIP &&
		!isDonor &&
		len(torrent.CategoryMode) != 0 {
		
			var ConfHrMod string
			if ConfHrMod == constant.HR_MODE_GLOBAL || 
			(ConfHrMod == constant.HR_MODE_MANUAL && torrent.HR == constant.HR_TORRENT_YES) {
				
			}


	}


	// VIP do not calculate downloaded
	if user.Class == constant.UC_VIP{
		
	}


	return
}

func Max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
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

func hashPad(hash string) string {
	return fmt.Sprintf("%-20s", string(hash))
}

func concatIPAndPort(ipv4 string, port uint16) []byte {
	ip := net.ParseIP(ipv4)
	if ip == nil {
		// 处理无效的 IPv4 地址
		return nil
	}

	ipv4Bytes := ip.To4()
	if ipv4Bytes == nil {
		// 处理非 IPv4 地址
		return nil
	}

	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, port)

	return append(ipv4Bytes, portBytes...)
}
