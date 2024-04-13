package biz

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"pt/internal/biz/constant"
	"pt/internal/biz/inter"
	"pt/internal/biz/model"
)

type AnnounceRequest struct {
	InfoHash      string `binding:"required" query:"info_hash" json:"info_hash,omitempty" bson:"info_hash" form:"info_hash"`
	PeerID        string `binding:"required" query:"peer_id" json:"peer_id,omitempty" bson:"peer_id" form:"peer_id"`
	IP            string `query:"ip" json:"ip,omitempty" bson:"ip" form:"ip"`
	Port          uint16 `binding:"required" query:"port" json:"port,omitempty" bson:"port" form:"port"`
	Uploaded      uint   `binding:"required" query:"uploaded" json:"uploaded,omitempty" bson:"uploaded" form:"uploaded"`
	Downloaded    uint   `binding:"required" query:"downloaded" json:"downloaded,omitempty" bson:"downloaded" form:"downloaded"`
	Left          uint   `binding:"required" query:"left" json:"left,omitempty" bson:"left" form:"left"`
	Numwant       uint   `query:"numwant" json:"numwant,omitempty" bson:"numwant" form:"numwant"` //TODO num want, num_want
	Key           string `query:"key" json:"key,omitempty" bson:"key" form:"key"`
	Compact       bool   `query:"compact" json:"compact,omitempty" bson:"compact" form:"compact"`
	SupportCrypto bool   `query:"supportcrypto" json:"support_crypto,omitempty" bson:"support_crypto" form:"support_crypto"`
	Event         string `query:"event" json:"event,omitempty" bson:"event" form:"event"`

	Passkey  string `json:"passkey,omitempty" bson:"passkey" form:"passkey"`
	Authkey  string `json:"authkey,omitempty" bson:"authkey" form:"authkey"`
	RawQuery string `json:"raw_query,omitempty" bson:"raw_query" form:"raw_query"`
}

func (req *AnnounceRequest) IsSeeding() bool {
	return req.Left == 0
}

type AnnounceResponse struct {
	Interval    int         `bencode:"interval"`
	MinInterval int         `bencode:"min interval"`
	Complete    int         `bencode:"complete"`
	Incomplete  int         `bencode:"incomplete"`
	Peers       interface{} `bencode:"peers"`
	PeersIPv6   interface{} `bencode:"peers_ipv6"`
}

// TrackerAnnounceRepo
// TrackerUsecase is a Tracker usecase.
type TrackerAnnounceUsecase struct {
	urepo  inter.UserRepo
	trepo  inter.TorrentRepo
	peerpo inter.PeerRepo
	cache  inter.CacheRepo
	snatch inter.SnatchedRepo
	log    *log.Helper
}

// NewTrackerAnnounceUsecase new a Tracker usecase.
func NewTrackerAnnounceUsecase(
	urepo inter.UserRepo,
	cache inter.CacheRepo,
	peerpo inter.PeerRepo,
	trepo inter.TorrentRepo,
	logger log.Logger) *TrackerAnnounceUsecase {
	return &TrackerAnnounceUsecase{
		urepo: urepo,
		trepo: trepo,
		cache: cache,
		log:   log.NewHelper(logger),
	}
}

var (
	ErrParamsInvalidInfoHash = errors.New("Invalid infohash")
	ErrParmsInvalidPassKey   = errors.New("Invalid passkey")
	ErrParamsInvalidAuthKey  = errors.New("authkey format error")
	ErrAgentBlock            = errors.New("Browser access blocked!")
)

type announceParamsChecker struct {
	AReq *AnnounceRequest
	Err  error
}

func NewAnnounceParamsChecker() *announceParamsChecker {
	return &announceParamsChecker{}
}

func (o *announceParamsChecker) Do(ctx http.Context) (*AnnounceRequest, error) {

	o.CheckUserAgent(ctx)
	o.Bind(ctx)

	return o.AReq, o.Err
}

// TODO checkclient
func (o *announceParamsChecker) CheckUserAgent(ctx http.Context) {
	if o.Err != nil {
		return
	}

	ua := ctx.Header().Get("User-Agent")
	patterns := []string{
		"^Mozilla/",
		"^Opera/",
		"^Links/",
		"^Lynx/",
	}

	for _, pattern := range patterns {
		match, err := regexp.MatchString(pattern, ua)
		if err != nil {
			o.Err = errors.WithStack(err)
			return
		}

		if match {
			o.Err = ErrAgentBlock
			return
		}
	}

}

func (o *announceParamsChecker) Bind(ctx http.Context) {
	if o.Err != nil {
		return
	}

	query := ctx.Request().URL.RawQuery
	req := new(AnnounceRequest)

	err := ctx.Bind(req)
	if err != nil {
		o.Err = err
		return
	}
	req.RawQuery = query
	o.AReq = req

	if len(req.InfoHash) != 20 {
		o.Err = ErrParamsInvalidInfoHash
	}

	if req.Passkey != "" && len(req.Passkey) != 32 {
		o.Err = ErrParmsInvalidPassKey
		return
	}

	if req.Authkey != "" {
		parts := strings.Split(req.Authkey, "|")
		if len(parts) != 3 {
			o.Err = ErrParamsInvalidAuthKey
			return
		}
	}
}

func (o *TrackerAnnounceUsecase) AnnounceParams(ctx http.Context) (*AnnounceRequest, error) {

	ac := NewAnnounceParamsChecker()

	return ac.Do(ctx)
}

type lockParam struct {
	User     string
	Infohash string
}

type lockPasskeyParam struct {
	Passkey  string
	Infohash string
}

// AnounceCheck, data check before response
func (o *TrackerAnnounceUsecase) AnounceHandler(ctx http.Context) (resp AnnounceResponse, err error) {

	// params check
	in, err := o.AnnounceParams(ctx)
	if err != nil {
		return
	}

	var (
		authKeyTid          string
		authKeyUid          string
		userAuthenticateKey string
		subAuthkey          string
		isReAnnounce        bool
	)

	authkey := in.Authkey
	passkey := in.Passkey
	infoHash := string(in.InfoHash[:])

	if authkey != "" {
		parts := strings.Split(authkey, "|")
		authKeyTid = parts[0]
		authKeyUid = parts[1]
		userAuthenticateKey = authKeyUid
		subAuthkey = fmt.Sprintf("%s|%s", authKeyTid, authKeyUid)

		// check ReAnnounce
		var isReAnnounce bool
		lockParams := &lockParam{
			User:     authKeyUid,
			Infohash: infoHash,
		}
		lockString := httpBuildQueryString(lockParams)
		lk := cacheKeyContactWithColon(constant.CacheKey_IsReAnnounceKey, string(md5.New().Sum([]byte(lockString))))
		success, err := o.cache.Lock(ctx, lk, time.Now().Unix(), 20)
		if err != nil {
			return resp, err
		}
		if !success { //false 键已存在 // 锁失败
			isReAnnounce = true
		}

		if !isReAnnounce {
			rcLock := cacheKeyContactWithColon(constant.CacheKey_ReAnnounceCheckByAuthKey, subAuthkey)
			success, err = o.cache.Lock(ctx, rcLock, time.Now().Unix(), 60)
			if err != nil {
				return resp, err
			}
			if !success { //false 键已存在
				return resp, errors.New("Request too frequent")
			}
		}

		authInvalidKey := cacheKeyContactWithColon(constant.CacheKey_AuthKeyInvalidKey, authkey)
		res, err := o.cache.Get(ctx, authInvalidKey)
		if err != nil {
			return resp, err
		}
		if len(res) != 0 {
			return resp, errors.New("Invalid authkey")
		}

	} else if passkey != "" {

		userAuthenticateKey = passkey

		pInkey := cacheKeyContactWithColon(constant.CacheKey_PasskeyInvalidKey, passkey)
		res, err := o.cache.Get(ctx, pInkey)
		if err != nil {
			return resp, err
		}
		if len(res) != 0 {
			return resp, errors.New("Passkey invalid")
		}

		lockParams := &lockPasskeyParam{
			Infohash: infoHash,
			Passkey:  passkey,
		}
		lockString := httpBuildPasskeyQueryString(lockParams)
		exist, err := o.cache.Lock(ctx, lockString, time.Now().Unix(), 20)
		if err != nil {
			return resp, err
		}
		if !exist { //false 键已存在
			isReAnnounce = true
		}

	} else {
		return resp, errors.New("Require passkey or authkey")
	}

	// 判断种子是否存在
	exist, err := o.cache.Get(ctx, cacheKeyContactWithColon(constant.CacheKey_TorrentNotExistsKey, infoHash))
	if err != nil {
		return
	}
	if len(exist) != 0 { //false 键已存在
		return resp, errors.New("Torrent not exists")
	}

	if !isReAnnounce {
		torrentReAnnounceKey := fmt.Sprintf("reAnnounceCheckByInfoHash:%s:%s", userAuthenticateKey, infoHash)

		exist, err := o.cache.Lock(ctx, torrentReAnnounceKey, time.Now().Unix(), 60)
		if err != nil {
			return resp, err
		}
		if !exist { //false 键已存在
			return resp, errors.New("Request too frequent")
		}
	}

	//dbconn_announce

	//check authkey //todo authKey not exist
	user, err := o.urepo.GetByAuthkey(ctx, authkey)
	if err != nil {
		key := cacheKeyContactWithColon(constant.CacheKey_AuthKeyInvalidKey, authkey)
		_, errLock := o.cache.Lock(ctx, key, time.Now().Unix(), 24*3600)
		if errLock != nil {
			return resp, errLock
		}
		return
	}
	passkey = user.Passkey //todo 这里会有全局passkey赋值

	var compact int //是否压缩
	ip := in.IP
	if in.Port > 0xffff {
		return resp, errors.New("invalid port")
	}
	p := net.ParseIP(ip) //Disable compact announce with IPv6
	if p == nil {
		compact = 0
	}

	var ipv4, ipv6 = "", ""
	if p != nil {
		if r := p.To4(); r != nil {
			ipv4 = ip
		} else {
			ipv6 = ip
		}
	}
	// TODO _GET["ipv4"]
	peerIPV46 := ""
	if ipv4 != "" {
		peerIPV46 += fmt.Sprintf(", ipv4 = %s", ipv4)
	}
	if ipv6 != "" {
		peerIPV46 += fmt.Sprintf(", ipv6 = %s", ipv6)
	}

	if portBlacklisted(in.Port) {
		return resp, errors.New(fmt.Sprintf("Port %d is blacklisted.", in.Port))
	}

	// return peer list limit
	rsize := 50 //TODO 暂时强制返回50条

	// seeder
	var seeder = "no"
	if in.Left == 0 {
		seeder = "yes"
	}

	o.log.Info(seeder, rsize)

	uInfoCacheKey := strFmtWithInsert(constant.CacheKey_UserPasskeyContent, passkey)
	azStr, err := o.cache.Get(ctx, uInfoCacheKey)
	if err != nil {
		return resp, err
	}
	if len(azStr) == 0 {
		user, err = o.urepo.GetByPasskey(ctx, passkey)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				o.cache.Lock(ctx, cacheKeyContactWithColon(constant.CacheKey_PasskeyInvalidKey, passkey), time.Now().Unix(), 24*3600)
				return resp, errors.New("Invalid passkey! Re-download the .torrent")
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

	// UserExist

	//IsDono
	var isDonor = isDono(user)

	//TODO: showclienterror

	// checkTorrent
	toData, err := o.cache.Get(ctx, strFmtWithInsert(constant.CacheKey_TorrentHashkeyContent, infoHash))
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
			infoHashUrlEncode := queryString[start:end] //TODO 这里为什么用encode,缓存会Miss
			o.log.Errorf("[TORRENT NOT EXISTS] params: %s, infoHashUrlEncode: %s\n", queryString, infoHashUrlEncode)
			o.cache.Set(ctx,
				cacheKeyContactWithColon(constant.CacheKey_TorrentNotExistsKey, infoHashUrlEncode),
				time.Now().Format(time.RFC3339), 24*3600)

			return resp, errors.Wrap(err, "torrent not registered with this tracker")
		}

		tobyte, _ := json.Marshal(torrent)
		toData = string(tobyte)
		o.cache.Set(ctx, strFmtWithInsert(constant.CacheKey_TorrentHashkeyContent, infoHash), toData, 350)
	}

	torrent := new(model.TorrentView)
	err = json.Unmarshal([]byte(toData), torrent)
	if err != nil {
		return resp, errors.Wrap(err, "Unmarshal")
	}

	if authKeyTid != "" && authKeyTid != strconv.FormatInt(int64(torrent.ID), 10) {
		err = errors.New("Invalid authkey")
		_, errT := o.cache.Lock(ctx, cacheKeyContactWithColon(constant.CacheKey_AuthKeyInvalidKey, authkey), time.Now().Unix(), 24*3600)
		if errT != nil {
			o.log.Errorw("cacheKeyContactWithColon")
		}
		return
	}

	if torrent.Banned == "yes" { //TODO 未判断用户权限
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
	peers, err := o.peerpo.GetPeerList(ctx, torrent.ID, onlyLeechQuery, limit)
	if err != nil {
		return
	}
	selfPeer := &model.PeerView{}
	if in.Event == "stop" {

	} else {
		var tmpPeerv4Bytes []byte
		var tmpPeerv6Bytes []byte

		var tmpPeerV4Bins []model.PeerBin
		var tmpPeerV6Bins []model.PeerBin

		for _, peer := range peers {
			peer.PeerID = hashPad(peer.PeerID)
			if peer.PeerID == in.PeerID && peer.UserID == user.Id {
				selfPeer = peer
				continue
			}

			if compact == 1 {
				if peer.IPv4 != "" {
					tmpPeerv4Bytes = append(tmpPeerv4Bytes, concatIPAndPort(peer.IPv4, peer.Port)...)
				}
				if peer.IPv6 != "" {
					tmpPeerv6Bytes = append(tmpPeerv6Bytes, concatIPAndPort(peer.IPv6, peer.Port)...)
				}
			} else {
				if peer.IPv4 != "" {
					tmpPeer := model.PeerBin{
						PeerID: peer.PeerID,
						IP:     peer.IPv4,
						Port:   int32(peer.Port),
					}
					tmpPeerV4Bins = append(tmpPeerV4Bins, tmpPeer)
				}

				if peer.IPv6 != "" {
					tmpPeer := model.PeerBin{
						PeerID: peer.PeerID,
						IP:     peer.IPv6,
						Port:   int32(peer.Port),
					}
					tmpPeerV6Bins = append(tmpPeerV6Bins, tmpPeer) //todo 这里也用Peers吗
				}
			}

		}

		if compact == 1 {
			resp.Peers = tmpPeerv4Bytes
			resp.PeersIPv6 = tmpPeerv6Bytes
		} else {
			resp.Peers = tmpPeerV4Bins
			resp.PeersIPv6 = tmpPeerV6Bins
		}
	}

	if selfPeer.ID == 0 {
		selfPeer, err = o.peerpo.GetPeerView(ctx, torrent.ID, user.Id, in.PeerID)
		if err != nil {
			return resp, err
		}
	}

	o.log.Infow("selfPeer", selfPeer)

	if selfPeer.ID != 0 && in.Event != "" && selfPeer.Prevts > (time.Now().Unix()-int64(announceWait)) {
		return resp, errors.New(fmt.Sprintf("There is a minimum announce time: %d wait", announceWait))
	}

	var isIPSeedBox int8 // TODO 盒子判断

	if constant.Setting_IsSeedBoxRuleEnabled {
		if ipv4 != "" {
			isIPSeedBox = 0
		}
	}
	o.log.Warn("isIPSeedBox", isIPSeedBox)

	var snatchInfo *model.Snatched
	var upthis, trueUpthis, downthis, trueDownthis int64
	var announcetime int64

	if selfPeer.ID == 0 {
		sameIPRecord, err := o.peerpo.GetPeer(ctx, torrent.ID, user.Id, ip)
		if err == nil && sameIPRecord.ID != 0 && seeder == "yes" {
			return resp, errors.New("You cannot seed the same torrent in the same location from more than 1 client.")
		}

		peers, err := o.peerpo.GetPeerListByUser(ctx, torrent.ID, user.Id)
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
						peerCnt, err := o.peerpo.CountPeersByUserAndSeedType(ctx, user.Id, "no")
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

		//buy torrent
		var flag = seeder == "no" &&
			user.SeedBonus != 0 &&
			torrent.Price > 0 &&
			torrent.Owner != user.Id &&
			constant.Setting_PaidTorrentEnabled

		if flag { //redisLock
			hasBuyCacheKey := cacheKeyContactWithColon(constant.CACHE_KEY_BOUGHT_USER_PREFIX, strconv.FormatInt(torrent.ID, 10))

			// hasBuy //consumeToBuyTorrent
			_, err := o.cache.HGet(ctx, hasBuyCacheKey, strconv.FormatInt(user.Id, 10))
			if err != nil {
				if errors.As(err, redis.Nil) { // 不存在

				}
				return resp, err
			}

		}
	} else { // continue an existing session

		upthis = Max(0, int64(in.Uploaded-uint(selfPeer.Uploaded)))
		trueUpthis = upthis
		downthis = Max(0, int64(in.Downloaded-uint(selfPeer.Downloaded)))
		trueDownthis = downthis
		var isCheater bool //todo
		o.log.Warn(isCheater)

		var seedTime int64
		var leechTime int64 // TODO: where can find

		o.log.Warn(announcetime)
		if selfPeer.Seeder == "yes" {
			announcetime = selfPeer.Announcetime + seedTime
		} else {
			announcetime = selfPeer.Announcetime + leechTime
		}

		if selfPeer.Announcetime > 0 && constant.Setting_IsSeedBoxRuleEnabled &&
			!(user.Class >= constant.UC_VIP || isDonor) && isIPSeedBox == 1 {

			// 获取速率设置项
			var notSeedBoxMaxSpeedMbps = 10 //TODO 获取速率设置项
			upSpeedMbps := calculateUpSpeedMbps(trueUpthis, selfPeer.Announcetime)
			if upSpeedMbps > float64(notSeedBoxMaxSpeedMbps) {
				//TODO 下载超速
				//updateDownloadPrivileges()
				//err
				return
			}

			//TODO checkCheater
			var isCheater bool
			//cheaterdet_security

			//snatchInfo
			snatchInfo, err = o.snatch.GetSnatched(ctx, torrent.ID, user.Id)
			if err != nil {
				return
			}
			if !isCheater && (trueUpthis > 0 || trueDownthis > 0) {
				//todo getDataTraffic
				o.log.Info(snatchInfo)
				// $dataTraffic = getDataTraffic($torrent, $_GET, $az, $self, $snatchInfo, apply_filter('torrent_promotion', $torrent));
				// $USERUPDATESET[] = "uploaded = uploaded + " . $dataTraffic['uploaded_increment_for_user'];
				// $USERUPDATESET[] = "downloaded = downloaded + " . $dataTraffic['downloaded_increment_for_user'];
			}

		}

	}

	// set non-type event
	var hasChangeSeederLeecher bool
	var times_completed_flag bool //标识该peer是否已经完成
	o.log.Warn(hasChangeSeederLeecher)
	var event string
	if selfPeer.ID != 0 && in.Event == "stopped" {
		affect, err := o.peerpo.Delete(ctx, selfPeer.ID)
		if err != nil {
			o.log.Error(err)
		} else { // todo updateSnatched
			if affect != 0 && snatchInfo != nil && snatchInfo.ID != 0 {
				hasChangeSeederLeecher = true
				err = o.snatch.UpdateSnatchedInfo(ctx, snatchInfo.ID, snatchInfo.Uploaded+trueUpthis, snatchInfo.Downloaded+trueDownthis, int64(in.Left))
				if err != nil {
					o.log.Errorw("err", err)
				}
			}
		}

	} else if selfPeer.ID != 0 {

		if in.Event == "completed" {
			updateMap := make(map[string]interface{})
			updateMap["ip"] = in.IP
			updateMap["port"] = in.Port
			updateMap["uploaded"] = in.Uploaded
			updateMap["downloaded"] = in.Downloaded
			updateMap["to_go"] = in.Left
			updateMap["prev_action"] = selfPeer.LastAction
			updateMap["last_action"] = time.Now().Format(time.DateOnly) //dt
			updateMap["seeder"] = seeder
			updateMap["agent"] = selfPeer.Agent
			updateMap["is_seed_box"] = isIPSeedBox
			updateMap["finished"] = "yes"
			updateMap["finishedat"] = time.Now().Unix()
			updateMap["completedat"] = time.Now().Format(time.DateOnly)
			_, err := o.peerpo.Update(ctx, selfPeer.ID, updateMap)
			if err != nil {
				o.log.Error("peerpo", err) //todo
			}
			times_completed_flag = true //todo 这里原来用的 $updateset[] = "times_completed = times_completed + 1";
			hasChangeSeederLeecher = true

			if err == nil {
				if seeder != selfPeer.Seeder {
					//  $updateset[] = ($seeder == "yes" ? "seeders = seeders + 1, leechers = leechers - 1" : "seeders = seeders - 1, leechers = leechers + 1");
					hasChangeSeederLeecher = true
				}
				if snatchInfo != nil && snatchInfo.ID != 0 {
					err := o.snatch.UpdateSnatchedInfo(ctx, snatchInfo.ID, trueUpthis, trueDownthis, int64(in.Left))
					if err != nil {
						o.log.Error("UpdateSnatchedInfo", err) //todo

					}
					//do_action('snatched_saved', $torrent, $snatchInfo); //todo
				}
			}
		}

	} else {

		if in.Event != "stopped" {

			//todo connectable 判断 ==
			var connectable = "false"

			// 插入peer
			newPeer := &model.Peer{
				Torrent:     torrent.ID,
				PeerID:      in.PeerID, //todo 验证
				IP:          ip,
				Port:        int64(in.Port),
				Uploaded:    int64(in.Uploaded),
				Downloaded:  int64(in.Downloaded),
				ToGo:        int64(in.Left),
				Seeder:      seeder,
				LastAction:  time.Now(),
				PrevAction:  time.Now(),
				Connectable: connectable,
				UserID:      user.Id,
				Agent:       "", //todo 获取agent
				Passkey:     passkey,
				IPv4:        ipv4,
				IPv6:        ipv6,
				IsSeedBox:   isIPSeedBox,
			}

			err = o.peerpo.Insert(ctx, newPeer)
			if err != nil {
				return resp, err
			}

			hasChangeSeederLeecher = true
			//$checkSnatchedRes = mysql_fetch_assoc(sql_query("SELECT id FROM snatched WHERE torrentid = $torrentid AND userid = $userid limit 1"));
			snatchInfo, err := o.snatch.GetSnatched(ctx, torrent.ID, user.Id)
			if err != nil && !errors.As(err, gorm.ErrRecordNotFound) {
				return resp, err
			}

			// 插入
			if snatchInfo == nil || snatchInfo.ID == 0 {
				err = o.snatch.Insert(ctx, &model.Snatched{
					Torrentid:  torrent.ID,
					UserID:     user.Id,
					IP:         ip,
					Port:       in.Port,
					Uploaded:   int64(in.Uploaded),
					Downloaded: int64(in.Downloaded),
					ToGo:       int64(in.Left),
					LastAction: time.Now(),
					StartAt:    time.Now(),
				})
				if err != nil {
					return resp, err
				}
			} else {
				//更新
				snatchMap := make(map[string]interface{}, 0)
				snatchMap["to_go"] = in.Left
				snatchMap["last_action"] = time.Now()
				err = o.snatch.UpdateWithMap(ctx, snatchInfo.ID, snatchMap)
				if err != nil {
					return resp, err
				}
			}

		}

	}

	//handle hr
	if in.Left > 0 || event == "completed" && user.Class < constant.UC_VIP && !isDonor && len(torrent.CategoryMode) != 0 {

		var ConfHrMod string
		if ConfHrMod == constant.HR_MODE_GLOBAL ||
			(ConfHrMod == constant.HR_MODE_MANUAL && torrent.HR == constant.HR_TORRENT_YES) {

		}

	}

	// update torrentInfo
	var (
		seederCnt, _  = o.peerpo.GetCountByTrackerState(ctx, torrent.ID, constant.Seeder)
		leecherCnt, _ = o.peerpo.GetCountByTrackerState(ctx, torrent.ID, constant.Leecher)
	)
	if hasChangeSeederLeecher {
		infoMap := make(map[string]interface{})
		infoMap["seeders"] = seederCnt
		infoMap["leechers"] = leecherCnt
		infoMap["variable"] = "yes"
		infoMap["last_action"] = time.Now()
		if times_completed_flag {
			infoMap["times_completed_flag"] = 1
		}
		err = o.trepo.UpdateByMap(ctx, torrent.ID, infoMap)
		if err != nil {
			o.log.Error("updateByMapErr", err)
		}
	}

	// 更新用户上传下载信息 USERUPDATESET
	// VIP do not calculate downloaded
	if user.Class == constant.UC_VIP {

	}

	return
}

func Max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func httpBuildQueryString(params *lockParam) string {
	return fmt.Sprintf("user=%s&info_hash=%s", params.User, params.Infohash)
}

func httpBuildPasskeyQueryString(params *lockPasskeyParam) string {
	return fmt.Sprintf("passkey=%s&info_hash=%s", params.Passkey, params.Infohash)
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

func cacheKeyContactWithColon(key, body string) string {
	return fmt.Sprintf("%s:%s", key, body)
}

func strFmtWithInsert(key, body string) string {
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

func isDono(u *model.User) bool {
	if u.Donor != "yes" {
		return false
	}
	if u.DonorUntil == nil ||
		u.DonorUntil.Compare(time.Now()) == 1 ||
		u.DonorUntil.Format("2006-01-02 15:04:05") == "0000-00-00 00:00:00" ||
		u.DonorUntil.Format("2006-01-02 15:04:05") == "0001-01-01 00:00:00" {
		return true
	}

	return false
}

func calculateUpSpeedMbps(trueupthis int64, announcetime int64) float64 {
	upSpeedMbps := ((float64(trueupthis) / float64(announcetime)) / 1024 / 1024) * 8
	upSpeedMbps = math.Round(upSpeedMbps*100) / 100 // 保留两位小数
	return upSpeedMbps
}
