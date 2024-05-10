package biz

import (
	"context"
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

	UA string `json:"ua,omitempty" bson:"ua" form:"ua"`
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
	urepo       inter.UserRepo
	trepo       inter.TorrentRepo
	peerpo      inter.PeerRepo
	cache       inter.CacheRepo
	snatch      inter.SnatchedRepo
	hitpo       inter.HitRunsRepo
	agentpo     inter.AgentAllowedRepo
	agentDenyPo inter.AgentDenyRepo
	cheaterRepo inter.CheaterRepo
	seedBoxRepo inter.SeedBoxRepo
	log         *log.Helper
}

// NewTrackerAnnounceUsecase new a Tracker usecase.
func NewTrackerAnnounceUsecase(
	urepo inter.UserRepo,
	cache inter.CacheRepo,
	peerpo inter.PeerRepo,
	trepo inter.TorrentRepo,
	snatch inter.SnatchedRepo,
	hitpo inter.HitRunsRepo,
	agentpo inter.AgentAllowedRepo,
	agentDenyPo inter.AgentDenyRepo,
	cheaterRepo inter.CheaterRepo,
	seedBoxRepo inter.SeedBoxRepo,
	logger log.Logger) *TrackerAnnounceUsecase {
	return &TrackerAnnounceUsecase{
		urepo:       urepo,
		trepo:       trepo,
		peerpo:      peerpo,
		cache:       cache,
		snatch:      snatch,
		hitpo:       hitpo,
		agentpo:     agentpo,
		agentDenyPo: agentDenyPo,
		cheaterRepo: cheaterRepo,
		seedBoxRepo: seedBoxRepo,
		log:         log.NewHelper(logger),
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
}

func NewAnnounceParamsChecker() *announceParamsChecker {

	return &announceParamsChecker{
		AReq: &AnnounceRequest{},
	}
}

func (o *announceParamsChecker) Do(ctx http.Context) (*AnnounceRequest, error) {
	err := o.CheckUserAgent(ctx)
	if err != nil {
		return nil, err
	}
	err = o.Bind(ctx)
	if err != nil {
		return nil, err
	}
	return o.AReq, nil
}

// checkclient
func (o *TrackerAnnounceUsecase) CheckClient(ctx http.Context, in *AnnounceRequest) error {
	allows := make([]*model.AgentAllowedFamily, 0)

	allowStr, err := o.cache.Get(ctx, constant.CacheKeyAgentAllowKey)
	if err != nil && err != redis.Nil {
		return errors.Wrap(err, "Get")
	}

	if len(allowStr) != 0 {
		json.Unmarshal([]byte(allowStr), &allows)
	} else {
		allows, err = o.agentpo.GetList(ctx)
		if err != nil {
			return errors.Wrap(err, "GetList")
		}
		data, _ := json.Marshal(allows)
		o.cache.Set(ctx, constant.CacheKeyAgentAllowKey, string(data), 60*60*24*time.Second)
	}

	var agentAllowPassed *model.AgentAllowedFamily
	var versionTooLowStr string

	for _, allow := range allows {

		var (
			isPeerIdAllowed = false
			isAgentAllowed  = false
			isPeerIdTooLow  = false
			isAgentTooLow   = false
		)

		if allow.PeerIdPattern == "" || in.PeerID == "" {
			isPeerIdAllowed = true
		} else {

			pattern := allow.PeerIdPattern
			start := allow.PeerIdStart
			matchType := allow.PeerIdMatchType
			matchNum := allow.PeerIdMatchNum

			//$peerIdResult = $this->isAllowed
			peerIdResult := o.isAllowed(pattern, start, in.PeerID, matchNum, matchType)
			if peerIdResult == 1 {
				isPeerIdAllowed = true
			}
			if peerIdResult == 2 {
				isPeerIdTooLow = true
			}
		}

		if isPeerIdAllowed && isAgentAllowed {
			agentAllowPassed = allow
			break
		}
		if isPeerIdTooLow && isAgentTooLow {

			//versionTooLowStr = "Your " . $agentAllow->family . " 's version is too low, please update it after " . $agentAllow->start_name;

			versionTooLowStr = fmt.Sprintf("Your %s 's version is too low,please update it after %s", allow.Family, allow.StartName)
			break //todo php代码里面没有
		}

	}

	if len(versionTooLowStr) != 0 {
		return errors.New(versionTooLowStr)
	}

	if agentAllowPassed == nil {
		//throw new ClientNotAllowedException("Banned Client, Please goto " . getSchemeAndHttpHost() . "/faq.php#id29 for a list of acceptable clients");
		return errors.New("Banned Client")
	}

	if agentAllowPassed.Exception == "Yes" {
		// check if exclude
		// checkIsDenied($peerId, $agent, $familyId)
		allowDenys := make([]*model.AgentAllowedException, 0)

		allowStr, err := o.cache.Get(ctx, constant.CacheKeyAgentDenyKey)
		if err != nil && err != redis.Nil {
			return errors.Wrap(err, "Get")
		}

		if len(allowStr) != 0 {
			json.Unmarshal([]byte(allowStr), &allowDenys)
		} else {
			allowDenys, err = o.agentDenyPo.GetList(ctx)
			if err != nil {
				return errors.Wrap(err, "GetList")
			}

			data, _ := json.Marshal(allowDenys)
			o.cache.Set(ctx, constant.CacheKeyAgentDenyKey, string(data), 60*60*24*time.Second)
		}

		allowDenysFiltered := make([]*model.AgentAllowedException, 0)
		for _, deny := range allowDenys {
			if deny.FamilyID != int(agentAllowPassed.ID) {
				continue
			}
			allowDenysFiltered = append(allowDenysFiltered, deny)
		}

		for _, agentDeny := range allowDenysFiltered {
			if agentDeny.Agent == in.UA && regexp.MustCompile("^"+agentDeny.PeerId).MatchString(in.PeerID) {
				return fmt.Errorf("[%d-%d]Client: %s is banned due to: %s", agentAllowPassed.ID, agentDeny.ID, agentDeny.Name, *agentDeny.Comment)
			}
		}

	}

	if ctx.Request().URL.Scheme == "https" && agentAllowPassed.Allowhttps != "Yes" {
		return errors.New("This client does not support https well")
	}

	return nil

}

func (o *TrackerAnnounceUsecase) getPatternMatches(pattern, start string, matchNum int64) ([]string, error) {
	matches := regexp.MustCompile(pattern).FindStringSubmatch(start)
	if matches == nil {
		return nil, fmt.Errorf("pattern: %s can not match start: %s", pattern, start)
	}

	matchCount := len(matches) - 1
	if matchNum > int64(matchCount) { // note: php 代码就这样
		// Remove this check for compatibility with legacy behavior.
		// return nil, fmt.Errorf("pattern: %s match start: %s got matches count: %d, but require %d.", pattern, start, matchCount, matchNum)
	}

	return matches[1 : matchNum+1], nil
}

func (o *TrackerAnnounceUsecase) isAllowed(pattern, start, value string, matchNum int64, matchType string) int {

	logPrefix := ""
	matchBench, _ := o.getPatternMatches(pattern, start, matchNum)
	o.log.Debugf("%s, matchBench: %v", logPrefix, matchBench)

	if !regexp.MustCompile(pattern).MatchString(value) {
		o.log.Errorf("%s, pattern: (%s) not match: (%s)", logPrefix, pattern, value)
		return 0
	}

	if matchNum <= 0 {
		return 1
	}

	matchTarget := regexp.MustCompile(pattern).FindStringSubmatch(value)[1:]
	o.log.Debugf("%s, matchTarget: %v", logPrefix, matchTarget)

	for i := 0; i < int(matchNum); i++ {
		if i >= len(matchBench) || i >= len(matchTarget) {
			break
		}

		var benchValue, targetValue int

		switch matchType {
		case "dec":
			benchValue = parseInt(matchBench[i], 10)
			targetValue = parseInt(matchTarget[i], 10)
		case "hex":
			benchValue = parseInt(matchBench[i], 16)
			targetValue = parseInt(matchTarget[i], 16)
		default:
			panic(fmt.Errorf("Invalid match type: %s", matchType))
		}

		if targetValue > benchValue {
			// Higher, pass directly
			return 1
		} else if targetValue < benchValue {
			return 2
		}
	}

	// NOTE: At last, after all positions checked, not [NOT_MATCH] or lower, it is passed!
	return 1
}

func parseInt(s string, base int) int {
	i, err := strconv.ParseInt(s, base, 64)
	if err != nil {
		panic(err)
	}
	return int(i)
}
func (o *announceParamsChecker) CheckUserAgent(ctx http.Context) error {

	req := new(AnnounceRequest)
	ua := ctx.Header().Get("User-Agent")
	req.UA = ua
	patterns := []string{
		"^Mozilla/",
		"^Opera/",
		"^Links/",
		"^Lynx/",
	}

	for _, pattern := range patterns {
		match, err := regexp.MatchString(pattern, ua)
		if err != nil {
			return errors.Wrap(err, "MatchString")
		}

		if match {
			return errors.New(ErrAgentBlock.Error())
		}
	}

	return nil
}

func (o *announceParamsChecker) Bind(ctx http.Context) error {

	query := ctx.Request().URL.RawQuery

	err := ctx.Bind(o.AReq)
	if err != nil {
		return errors.Wrap(err, "BindQuery")
	}
	o.AReq.RawQuery = query

	if len(o.AReq.InfoHash) != 20 {
		return errors.New(ErrParamsInvalidInfoHash.Error())
	}

	if o.AReq.Passkey != "" && len(o.AReq.Passkey) != 32 {
		return errors.New(ErrParmsInvalidPassKey.Error())
	}

	if o.AReq.Authkey != "" {
		parts := strings.Split(o.AReq.Authkey, "|")
		if len(parts) != 3 {
			return errors.New(ErrParamsInvalidAuthKey.Error())
		}
	}
	return nil
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
	o.CheckClient(ctx, in)

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
	rsize := 50 //TODO 暂时强制最多返回50条

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

	var isIPSeedBoxInt int8 // 盒子判断
	if constant.Setting_IsSeedBoxRuleEnabled {

		var isIPSeedBox bool
		if ipv4 != "" {
			isIPSeedBox, err = o.isIPSeedBox(ctx, ipv4, user.Id)
			if err != nil {
				return
			}
		}

		if !isIPSeedBox && ipv6 != "" {
			isIPSeedBox, err = o.isIPSeedBox(ctx, ipv6, user.Id)
			if err != nil {
				return
			}
		}
		if isIPSeedBox {
			isIPSeedBoxInt = 1
		}

	}
	o.log.Warn("isIPSeedBoxInt", isIPSeedBoxInt)

	var snatchInfo *model.Snatched
	var upthis, trueUpthis, downthis, trueDownthis int64
	//var announcetime int64 //无用

	var dataTraffic = make(map[string]int64, 0)

	if selfPeer.ID == 0 {
		sameIPRecord, err := o.peerpo.GetPeer(ctx, torrent.ID, user.Id, ip, "")
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

			var wait int64
			if constant.WaitSystem == "yes" {
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
			if constant.MaxDlSystem == "yes" {
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
			hasBuyCacheKey := getTorrentBuyUserCacheKey(torrent.ID)

			// hasBuy //consumeToBuyTorrent
			hasBuy, err := o.cache.HGetBool(ctx, hasBuyCacheKey, strconv.FormatInt(user.Id, 10))
			if err != nil {
				if !errors.As(err, redis.Nil) {
					return resp, err
				}
				// 刷新缓存
				lockKey := getTorrentBuyLoadUserLockKey(user.Id)
				ok, _ := o.cache.Lock(ctx, lockKey, "1", 300*time.Millisecond) // todo 这里不会等待
				if ok {
					//executeCommand("torrent:load_bought_user $torrentid", "string", true, false);
					err = o.refreshHasBuyLogs(ctx, torrent.ID)
					if err != nil {
						return resp, err
					}
				}
				// note: 这里将原有string,改为hget
				// $hasBuy = \Nexus\Database\NexusDB::remember(sprintf("user_has_buy_torrent:%s:%s", $userid, $torrentid), 86400*10, function () use($userid, $torrentid) {
				// 	$exists = \App\Models\TorrentBuyLog::query()->where('uid', $userid)->where('torrent_id', $torrentid)->exists();
				// 	return intval($exists);
				// });
				hasBuy, _ = o.cache.HGetBool(ctx, hasBuyCacheKey, strconv.FormatInt(user.Id, 10))
			}

			if !hasBuy {

				lockKey := fmt.Sprintf("buying_torrent:%d", user.Id)
				_, err = o.cache.Lock(ctx, lockKey, "1", 5*time.Second) //TODO 5s?  $lock = new \Nexus\Database\NexusLock("", 5);
				if err != nil {
					return resp, errors.New("buying torrent, wait!")
				}

				_, err := o.consumeToBuyTorrent(ctx, user.Id, torrent.ID, "Web")
				if err != nil {
					return resp, err
				}
				o.cache.HSet(ctx, hasBuyCacheKey, strconv.FormatInt(user.Id, 10), true)
				//$lock->release();
			}

		}
	} else { // continue an existing session

		upthis = Max(0, int64(in.Uploaded-uint(selfPeer.Uploaded)))
		trueUpthis = upthis
		downthis = Max(0, int64(in.Downloaded-uint(selfPeer.Downloaded)))
		trueDownthis = downthis

		// var seedTime int64
		// var leechTime int64
		// if selfPeer.Seeder == "yes" {
		// 	announcetime = selfPeer.Announcetime + seedTime
		// } else {
		// 	announcetime = selfPeer.Announcetime + leechTime
		// }
		//o.log.Warn(announcetime) //TODO snatch表无此字段,就不更新了

		if selfPeer.Announcetime > 0 && constant.Setting_IsSeedBoxRuleEnabled &&
			!(user.Class >= constant.UC_VIP || isDonor) && isIPSeedBoxInt == 1 {

			// 获取速率设置项
			upSpeedMbps := calculateUpSpeedMbps(trueUpthis, selfPeer.Announcetime)
			if upSpeedMbps > float64(constant.NotSeedBoxMaxSpeedMbps) { //超速
				//updateDownloadPrivileges()
				user.UploadPos = "no"
				o.urepo.Update(ctx, user)
				o.clearUserCache(ctx, user.Id, passkey)
				return
			}
		}

		var isCheater bool
		//cheaterdet_security
		if constant.CheateredSecurity > 0 {
			if user.Class < constant.NoDetectSecurityUserClass && selfPeer.Announcetime > 10 {
				isCheater, err = o.cheaterCheck(ctx, user, torrent.ID, upthis, downthis, selfPeer.Announcetime, torrent.Seeders, torrent.Leechers)
				if err != nil {
					return resp, err
				}
			}
		}

		//snatchInfo
		snatchInfo, err = o.snatch.GetSnatched(ctx, torrent.ID, user.Id)
		if err != nil {
			return
		}
		if !isCheater && (trueUpthis > 0 || trueDownthis > 0) {
			o.log.Info(snatchInfo)
			// $dataTraffic = getDataTraffic($torrent, $_GET, $az, $self, $snatchInfo, apply_filter('torrent_promotion', $torrent));
			dataTraffic, err = o.getDataTraffic(ctx, torrent, in, user, selfPeer, snatchInfo, nil)
			if err != nil {
				return
			}

			// 在下面实现
			// $USERUPDATESET[] = "uploaded = uploaded + " . $dataTraffic['uploaded_increment_for_user'];
			// $USERUPDATESET[] = "downloaded = downloaded + " . $dataTraffic['downloaded_increment_for_user'];
		}
	}

	// set non-type event
	var hasChangeSeederLeecher bool
	var times_completed_flag bool //标识该peer是否已经完成
	var event string
	if selfPeer.ID != 0 && in.Event == "stopped" {
		affect, err := o.peerpo.Delete(ctx, selfPeer.ID)
		if err != nil {
			o.log.Error(err)
		} else {
			if affect != 0 && snatchInfo != nil && snatchInfo.ID != 0 {
				hasChangeSeederLeecher = true
				err = o.snatch.UpdateSnatchedInfo(ctx, snatchInfo.ID, trueUpthis, trueDownthis, int64(in.Left))
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
			updateMap["agent"] = in.UA
			updateMap["is_seed_box"] = isIPSeedBoxInt
			updateMap["finished"] = "yes"
			updateMap["finishedat"] = time.Now().Unix()
			updateMap["completedat"] = time.Now().Format(time.DateOnly)
			_, err := o.peerpo.Update(ctx, selfPeer.ID, updateMap)
			if err != nil {
				o.log.Error("peerpo", err)
			}
			times_completed_flag = true
			hasChangeSeederLeecher = true

			if err == nil {
				if seeder != selfPeer.Seeder {
					//  $updateset[] = ($seeder == "yes" ? "seeders = seeders + 1, leechers = leechers - 1" : "seeders = seeders - 1, leechers = leechers + 1");
					hasChangeSeederLeecher = true
				}
				if snatchInfo != nil && snatchInfo.ID != 0 {
					err := o.snatch.UpdateSnatchedInfo(ctx, snatchInfo.ID, trueUpthis, trueDownthis, int64(in.Left))
					if err != nil {
						o.log.Error("UpdateSnatchedInfo", err)
					}
					//do_action('snatched_saved', $torrent, $snatchInfo); //TODO 这里没找到php function在哪里 就没实现
				}
			}
		}

	} else {

		if in.Event != "stopped" {
			peer, err := o.peerpo.GetPeer(ctx, torrent.ID, user.Id, "", in.PeerID)
			if err != nil && !errors.As(err, gorm.ErrRecordNotFound) {
				return resp, err
			}
			if peer == nil || peer.ID == 0 {

				//TODO connectable 判断 ==  这个php源码会建立一条socket连接,但实际查询PeerList时没任何作用,暂不实现
				var connectable = "false"

				// 插入peer
				newPeer := &model.Peer{
					Torrent:     torrent.ID,
					PeerID:      in.PeerID,
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
					Agent:       in.UA,
					Passkey:     passkey,
					IPv4:        ipv4,
					IPv6:        ipv6,
					IsSeedBox:   isIPSeedBoxInt,
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
	}

	//handle hr
	if in.Left > 0 || event == "completed" && user.Class < constant.UC_VIP && !isDonor && len(torrent.Mode) != 0 {

		var ConfHrMod string
		if ConfHrMod == constant.HR_MODE_GLOBAL || (ConfHrMod == constant.HR_MODE_MANUAL && torrent.HR == constant.HR_TORRENT_YES) {
			hrCacheKey := getCacheKeyHR(user.Id, torrent.ID) //todo 时间够了需要删除hr状态
			hrCacheResult, err := o.cache.Get(ctx, hrCacheKey)
			if err != nil && err != redis.Nil {
				return resp, err
			}
			hrCacheNotExists := len(hrCacheResult) == 0
			if hrCacheNotExists {

				// 获取mysql
				_, err := o.hitpo.Get(ctx, torrent.ID, user.Id)
				if err == nil { // 存入缓存
					o.cache.Set(ctx, hrCacheKey, "hr", 24*3600*time.Second) // todo set value值是什么?
				} else if !errors.As(err, gorm.ErrRecordNotFound) {
					return resp, err
				} else {
					// 需创建
					includeRate := o.getIncludeRateByTorrentMode(torrent.Mode)
					// get newest snatch info
					snatchInfo, err := o.snatch.GetSnatched(ctx, torrent.ID, user.Id)
					if err != nil {
						o.log.Error("snatchInfo", err)
						return resp, err
					}
					requiredDownloaded := int64(float64(torrent.Size) * includeRate)
					if snatchInfo.Downloaded >= requiredDownloaded {
						hr := &model.HitRuns{
							UID:        user.Id,
							TorrentID:  torrent.ID,
							SnatchedID: snatchInfo.ID,
							Status:     0,
							Comment:    "",
							CreatedAt:  time.Now(),
							UpdatedAt:  time.Now(),
						}

						err = o.hitpo.Create(ctx, hr)
						if err != nil {
							o.log.Error("hitpo.Create", err)
						}
						o.cache.Set(ctx, hrCacheKey, "hr", 24*3600*time.Second)
					}
				}

			}
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

	// clientFamilyId
	var clientFamilyId int64
	var clientSelect int64
	if clientFamilyId != 0 && clientFamilyId != user.ClientSelect {
		clientSelect = clientFamilyId
	}

	//更新用户上传下载信息
	// VIP do not calculate downloaded
	if user.Class == constant.UC_VIP {
		// delete带download的数据 如 downloaded_increment_for_user
	}

	// 详细数据见 nexsusPHP中USERUPDATESET
	user.Uploaded += int64(dataTraffic["uploaded_increment_for_user"])
	if user.Class == constant.UC_VIP {
		// VIP do not calculate downloaded
	} else {
		user.Downloaded += int64(dataTraffic["downloaded_increment_for_user"])
	}
	if clientSelect != 0 {
		user.ClientSelect = clientSelect
	}
	err = o.urepo.Update(ctx, user)

	return
}

func (o *TrackerAnnounceUsecase) clearUserCache(ctx context.Context, uid int64, passkey string) {

	delKeys := []string{
		fmt.Sprintf("user_%d_content", uid),
		fmt.Sprintf("user_%d_roles", uid),
		fmt.Sprintf("announce_user_passkey__%d", uid),
		fmt.Sprintf("nexus_direct_permissions_%d", uid),
		fmt.Sprintf("user_role_ids:%d", uid),
		fmt.Sprintf("direct_permissions:%d", uid),
	}
	if passkey != "" {
		delKeys = append(delKeys, fmt.Sprintf("user_passkey_%s_content", passkey))
	}

	o.cache.Del(ctx, delKeys)
}

func (o *TrackerAnnounceUsecase) clearInboxCountCache(ctx context.Context, uid int64) {
	delKeys := []string{
		fmt.Sprintf("user_%d_inbox_count", uid),
		fmt.Sprintf("user_%d_unread_message_count", uid),
	}

	o.cache.Del(ctx, delKeys)
}

func (o *TrackerAnnounceUsecase) cheaterCheck(ctx context.Context, user *model.User, torrentid, uploaded, downloaded, anctime, seeders, leechers int64) (bool, error) {

	var upspeed = int64(0)
	if uploaded > 0 {
		upspeed = uploaded / anctime
	}

	//$mustBeCheaterSpeed = 1024 * 1024 * 1000; //1000 MB/s
	mustBeCheaterSpeed := constant.MaximumUploadSpeed * 1024 * 1024 / 8
	//$mayBeCheaterSpeed = 1024 * 1024 * 100; //100 MB/s
	mayBeCheaterSpeed := mustBeCheaterSpeed / 2

	if uploaded > constant.TrafficCntPerG && upspeed > (int64(mustBeCheaterSpeed)/constant.CheateredSecurity) {
		//Uploaded more than 1 GB with uploading rate higher than 100 MByte/S (For Consertive level). This is no doubt cheating.
		comment := fmt.Sprintf("User account was automatically disabled by system (Upload speed: %d MB/s)", uploaded/1024/1024/8)
		cheater := &model.Cheaters{
			ID:         0,
			Added:      time.Now(),
			UserID:     user.Id,
			TorrentID:  torrentid,
			Uploaded:   uploaded,
			Downloaded: downloaded,
			Anctime:    anctime,
			Seeders:    seeders,
			Leechers:   leechers,
			Comment:    comment,
		}
		err := o.cheaterRepo.Create(ctx, cheater)
		if err != nil {
			return true, err
		}
		user.Enabled = "no"
		err = o.urepo.Update(ctx, user)
		if err != nil {
			return true, err
		}

		return true, nil
	}

	if uploaded > constant.TrafficCntPerG && upspeed > int64(mayBeCheaterSpeed/constant.CheateredSecurity) {
		//Uploaded more than 1 GB with uploading rate higher than 25 MByte/S (For Consertive level). This is likely cheating.
		startTime := time.Now().AddDate(0, 0, -1)
		cheaterList, err := o.cheaterRepo.Query(ctx, user.Id, torrentid, startTime)
		if err != nil {
			return false, err
		}
		if len(cheaterList) == 0 {
			comment := "User is uploading fast when there is few leechers"
			cheater := &model.Cheaters{
				ID:         0,
				Added:      time.Now(),
				UserID:     user.Id,
				TorrentID:  torrentid,
				Uploaded:   uploaded,
				Downloaded: downloaded,
				Anctime:    anctime,
				Seeders:    seeders,
				Leechers:   leechers,
				Hit:        1,
				Comment:    comment,
			}
			err := o.cheaterRepo.Create(ctx, cheater)
			if err != nil {
				return false, err
			}

		} else {
			cheater := cheaterList[0]
			cheater.Hit++
			err := o.cheaterRepo.Update(ctx, cheater)
			if err != nil {
				return false, err
			}
		}
		return false, nil
	}

	if constant.CheateredSecurity <= 1 {
		return false, nil
	}

	if uploaded > constant.TrafficCntPerG && upspeed > constant.TrafficCntPerMB && leechers < (2*constant.CheateredSecurity) {
		startTime := time.Now().AddDate(0, 0, -1)
		cheaterList, err := o.cheaterRepo.Query(ctx, user.Id, torrentid, startTime)
		if err != nil {
			return false, err
		}
		if len(cheaterList) == 0 {
			comment := "User is uploading fast when there is few leechers"
			cheater := &model.Cheaters{
				ID:         0,
				Added:      time.Now(),
				UserID:     user.Id,
				TorrentID:  torrentid,
				Uploaded:   uploaded,
				Downloaded: downloaded,
				Anctime:    anctime,
				Seeders:    seeders,
				Leechers:   leechers,
				Hit:        1,
				Comment:    comment,
			}
			err := o.cheaterRepo.Create(ctx, cheater)
			if err != nil {
				return false, err
			}

		} else {
			cheater := cheaterList[0]
			cheater.Hit++
			err := o.cheaterRepo.Update(ctx, cheater)
			if err != nil {
				return false, err
			}
		}

		return false, nil
	}

	if uploaded > constant.TrafficCnt10MB && upspeed > constant.TrafficCnt100KB && leechers == 0 {
		//Uploaded more than 10 MB with uploading speed faster than 100 KByte/S when there is no leecher. This is likely cheating.
		startTime := time.Now().AddDate(0, 0, -1)
		cheaterList, err := o.cheaterRepo.Query(ctx, user.Id, torrentid, startTime)
		if err != nil {
			return false, err
		}
		if len(cheaterList) == 0 {
			comment := "User is uploading when there is no leecher"
			cheater := &model.Cheaters{
				ID:         0,
				Added:      time.Now(),
				UserID:     user.Id,
				TorrentID:  torrentid,
				Uploaded:   uploaded,
				Downloaded: downloaded,
				Anctime:    anctime,
				Seeders:    seeders,
				Leechers:   leechers,
				Hit:        1,
				Comment:    comment,
			}
			err := o.cheaterRepo.Create(ctx, cheater)
			if err != nil {
				return false, err
			}

		} else {
			cheater := cheaterList[0]
			cheater.Hit++
			err := o.cheaterRepo.Update(ctx, cheater)
			if err != nil {
				return false, err
			}
		}

		return false, nil

	}

	return false, nil
}

func (o *TrackerAnnounceUsecase) consumeToBuyTorrent(ctx context.Context, uid, torrentId int64, channel string) (bool, error) {

	if constant.TaxFactor < 0 || constant.TaxFactor > 1 {
		return false, errors.New(fmt.Sprintf("Invalid tax_factor: %f", constant.TaxFactor))
	}

	torrent, err := o.trepo.Get(ctx, torrentId)
	if err != nil {
		return false, err
	}
	requireBonus := torrent.Price

	user, err := o.urepo.Get(ctx, uid)
	if err != nil {
		return false, err
	}

	err = o.consumeUserBonus(ctx, user, requireBonus, constant.BUSINESS_TYPE_BUY_TORRENT)
	if err != nil {
		return false, err
	}
	o.trepo.CreateBuyLog(ctx, &model.TorrentBuyLog{
		UID:       uid,
		TorrentID: torrentId,
		Price:     requireBonus,
		Channel:   channel,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	increaseBonus := float64(requireBonus) * (1 - constant.TaxFactor)
	ownner, err := o.urepo.Get(ctx, torrent.Owner)
	if err != nil {
		return false, err
	}
	if ownner.Id != 0 {
		oldBonus := ownner.SeedBonus
		ownner.SeedBonus += increaseBonus
		o.urepo.Update(ctx, ownner)
		o.urepo.CreateBonusLog(ctx, &model.BonusLog{
			BusinessType:  constant.BUSINESS_TYPE_TORRENT_BE_DOWNLOADED,
			UID:           uid,
			OldTotalValue: oldBonus,
			Value:         increaseBonus,
			NewTotalValue: ownner.SeedBonus,
			// 'comment' => sprintf('[%s] %s', BonusLogs::$businessTypes[$businessType]['text'], $comment),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})

	}
	//TODO NOTE 未实现创建消息提醒: buyTorrentSuccessMessage
	//o.clearInboxCountCache(ctx, ownner.Id)
	return true, nil
}

func (o *TrackerAnnounceUsecase) consumeUserBonus(ctx context.Context, user *model.User, requireBonus int64, bussinessType int64) error {
	if requireBonus <= 0 {
		return nil
	}

	if user.SeedBonus < float64(requireBonus) {
		return errors.New("User bonus not enough.")
	}

	newUserBonus := user.SeedBonus - float64(requireBonus)
	user.SeedBonus = newUserBonus
	err := o.urepo.Update(ctx, user) // 更新用户信息
	if err != nil {
		return err
	}

	o.clearUserCache(ctx, user.Id, user.Passkey)
	return nil
}

func (o *TrackerAnnounceUsecase) getDataTraffic(ctx context.Context, torrent *model.TorrentView, queries *AnnounceRequest, user *model.User, peer *model.PeerView, snatch *model.Snatched, promotionInfo map[string]int64) (map[string]int64, error) {

	//!isset($user['__is_donor'] //return error

	var (
		realUploaded   = int64(0)
		realDownloaded = int64(0)
		downRatio      = float64(0)
		upRatio        = float64(0)
	)
	if peer == nil {
		realUploaded = int64(queries.Uploaded)
		realDownloaded = int64(queries.Downloaded)
		/**
		 * If peer not exits, user increment = 0;
		 */
		// $upRatio = 0;
		// $downRatio = 0;
	}

	if peer != nil {
		realUploaded = Max(int64(queries.Uploaded)-peer.Uploaded, 0)
		realDownloaded = Max(int64(queries.Downloaded)-peer.Downloaded, 0)
		// 全局优惠状态
		spStateReal := constant.TORRENT_PROMOTION_DEFAULT
		_, ok := promotionInfo["__ignore_global_sp_state"]
		if ok {
			spStateReal = promotionInfo["sp_state"]
		} else if constant.TORRENT_PROMOTION_GLOBAL != constant.TORRENT_PROMOTION_NORMAL {
			spStateReal = constant.TORRENT_PROMOTION_GLOBAL
		} else {
			spStateReal = torrent.SPState
		}
		if spStateReal == constant.TORRENT_PROMOTION_DEFAULT {
			spStateReal = constant.TORRENT_PROMOTION_NORMAL
		}

		uploaderRatio := constant.TorrentUploaderdouble

		if torrent.Owner == user.Id {
			//uploader, use the bigger one  // [IS_UPLOADER], upRatio: $upRatio";
			upRatio = math.Max(float64(uploaderRatio), constant.TorrentPromotionTypes[spStateReal].UpMultiplier)
		} else {
			//[IS_NOT_UPLOADER], upRatio: $upRatio";
			upRatio = constant.TorrentPromotionTypes[spStateReal].UpMultiplier
		}

		/**
		 * VIP do not calculate downloaded
		 * @since 1.7.13
		 */

		if user.Class == constant.UC_VIP {
			//$log .= ", [IS_VIP], downRatio: $downRatio";
			downRatio = 0
		} else {
			//$log .= ", [IS_NOT_VIP], downRatio: $downRatio";
			downRatio = constant.TorrentPromotionTypes[spStateReal].DownMultiplier
		}

	}

	uploadedIncrementForUser := int64(float64(realUploaded) * upRatio)
	downloadedIncrementForUser := int64(float64(realDownloaded) * downRatio)

	isSeedBoxRuleEnabled := constant.Setting_IsSeedBoxRuleEnabled
	if isSeedBoxRuleEnabled && torrent.Owner != user.Id && user.Class < constant.UC_VIP && !isDono(user) {

		is, err := o.isIPSeedBox(ctx, queries.IP, user.Id)
		if err != nil {
			return nil, err
		}
		if is {
			if constant.IsSeedBoxNoPromotion {
				uploadedIncrementForUser = realUploaded
				downloadedIncrementForUser = realDownloaded
				//$log .= ", isIPSeedBox && isSeedBoxNoPromotion, increment for user = real";
			}

			maxUploadedTimes := constant.MaxUploaded
			maxUploadedDurationSeconds := constant.MaxUploadedDuration * 3600
			torrentTTL := int64(time.Now().Sub(torrent.Added).Seconds()) //time() - strtotime($torrent['added']);
			timeRangeValid := (maxUploadedDurationSeconds == 0) || (torrentTTL < int64(maxUploadedDurationSeconds))
			//$log .= ", maxUploadedTimes: $maxUploadedTimes, maxUploadedDurationSeconds: $maxUploadedDurationSeconds, timeRangeValid: $timeRangeValid";
			if maxUploadedTimes > 0 && timeRangeValid {
				//$log .= ", [LIMIT_UPLOADED]";
				if snatch != nil && torrent.Size != 0 && snatch.Uploaded >= torrent.Size*int64(maxUploadedTimes) {
					//$log .= ", snatchUploaded({$snatch['uploaded']}) >= torrentSize({$torrent['size']}) * times($maxUploadedTimes), uploadedIncrementForUser = 0";
					uploadedIncrementForUser = 0
				} else {
					// $log .= ", snatchUploaded({$snatch['uploaded']}) < torrentSize({$torrent['size']}) * times($maxUploadedTimes), uploadedIncrementForUser do not change to 0";
				}
			} else {
				//$log .= ", [NOT_LIMIT_UPLOADED]";
			}

		}
	}

	resp := map[string]int64{
		"uploaded_increment":            realUploaded,
		"uploaded_increment_for_user":   uploadedIncrementForUser,
		"downloaded_increment":          realDownloaded,
		"downloaded_increment_for_user": downloadedIncrementForUser,
	}

	return resp, nil
}

func (o *TrackerAnnounceUsecase) isIPSeedBox(ctx context.Context, ip string, uid int64) (bool, error) {
	key := getIpSeedBoxCacheKey(ip, uid)
	result, err := o.cache.Get(ctx, key)
	if err != nil && err != redis.Nil {
		return false, err
	}
	if result == "1" {
		return true, nil
	}

	if result == "0" {
		return false, nil
	}

	ipNumeric, ipVersion, err := ipNumericAndVersion(ip) //这里ipv6与php计算有出入
	if err != nil {
		return false, err
	}

	seedBoxQ := &model.SeedBoxRecord{
		Status:         constant.SEED_BOX_STATUS_ALLOWED,
		IPBeginNumeric: ipNumeric,
		IPEndNumeric:   ipNumeric,
		Version:        ipVersion,
		IsAllowed:      1,
	}
	seedBox, err := o.seedBoxRepo.Query(ctx, seedBoxQ)
	if err != nil && !errors.As(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	if seedBox != nil && seedBox.ID != 0 {
		//do_log("$key, get result from database, is_allowed = 1, false");
		o.cache.Set(ctx, key, "0", 300*time.Second)
		return false, nil
	}

	seedBoxQ.IsAllowed = 0
	seedBoxQ.Type = constant.SEED_BOX_TYPE_ADMIN
	seedBox, err = o.seedBoxRepo.Query(ctx, seedBoxQ)
	if err != nil && !errors.As(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	if seedBox != nil && seedBox.ID != 0 {
		//do_log("$key, get result from admin, true");
		o.cache.Set(ctx, key, "1", 300*time.Second)
		return true, nil
	}

	if uid != 0 {
		seedBoxQ.UID = uid
		seedBoxQ.Type = constant.SEED_BOX_TYPE_USER
		seedBox, err = o.seedBoxRepo.Query(ctx, seedBoxQ)
		if err != nil && !errors.As(err, gorm.ErrRecordNotFound) {
			return false, err
		}
		if seedBox != nil && seedBox.ID != 0 {
			//do_log("$key, get result from user, true");
			o.cache.Set(ctx, key, "1", 300*time.Second)
			return true, nil
		}
	}

	//do_log("$key, no result, false");
	o.cache.Set(ctx, key, "0", 300*time.Second)

	return false, nil
}

func getIpSeedBoxCacheKey(ip string, uid int64) string {
	return fmt.Sprintf(constant.CacheKeyIpSeedBox, ip, uid)
}

func ipNumericAndVersion(ipStr string) (string, int64, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", 0, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	ipv4 := ip.To4()
	if ipv4 != nil {
		return strconv.FormatUint(uint64(binary.BigEndian.Uint32(ipv4)), 10), 4, nil
	}

	ipv6 := ip.To16()
	if ipv6 != nil {
		high := binary.BigEndian.Uint64(ipv6[:8])

		low := binary.BigEndian.Uint64(ipv6[8:])
		lowStr := strconv.FormatUint(low, 16)
		// 补齐至16位
		need := 16 - len(lowStr)
		add := ""
		for i := 0; i < need; i++ {
			add += "0"
		}

		return strconv.FormatUint(high, 16) + add + lowStr, 6, nil
	}

	return "", 0, errors.New("err format ip: " + ipStr)

}

func Max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func (o *TrackerAnnounceUsecase) refreshHasBuyLogs(ctx context.Context, tid int64) error {
	logs, err := o.trepo.GetBuyLogs(ctx, int64(tid))
	if err != nil {
		return err
	}
	hasBuyCacheKey := getTorrentBuyUserCacheKey(int64(tid))

	for _, log := range logs {
		o.cache.HSet(ctx, hasBuyCacheKey, strconv.FormatInt(int64(log.UID), 10), true)
	}

	o.cache.Expire(ctx, hasBuyCacheKey, time.Second*86400*30)

	return nil
}

// todo
func (o *TrackerAnnounceUsecase) getIncludeRateByTorrentMode(mode string) float64 {
	return 1
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

func getTorrentBuyLoadUserLockKey(uid int64) string {
	return fmt.Sprintf(constant.LOCK_KEY_LOAD_TORRENT_BOUGHT, uid)
}

func getTorrentBuyUserCacheKey(tid int64) string {
	return fmt.Sprintf(constant.CACHE_KEY_BOUGHT_USER_PREFIX, tid)
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

func getCacheKeyHR(userId, torrentId int64) string {
	return fmt.Sprintf(constant.CacheKey_HR, userId, torrentId)
}
