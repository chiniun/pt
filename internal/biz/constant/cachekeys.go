package constant

const CACHE_KEY_BOUGHT_USER_PREFIX = "torrent_purchasers"

const CACHE_KEY_PIECES_HASH = "torrent_pieces_hash"

const (
	CacheKey_TorrentNotExistsKey       = "torrent_not_exists"
	CacheKey_AuthKeyInvalidKey         = "authkey_invalid"
	CacheKey_PasskeyInvalidKey         = "passkey_invalid"
	CacheKey_IsReAnnounceKey           = "isReAnnounce"
	CacheKey_ReAnnounceCheckByAuthKey  = "reAnnounceCheckByAuthKey"
	CacheKey_ReAnnounceCheckByInfoHash = "reAnnounceCheckByInfoHash"

	CacheKey_UserPasskeyContent    = "user_passkey_%s_content"
	CacheKey_TorrentHashkeyContent = "torrent_hash_%s_content"

	CacheKey_HR = "hit_and_run:%d:%d" // {userId}{torrentId}

	// agentAllowed
	CacheKeyAgentAllowKey = "all_agent_allows:php"
	CacheKeyAgentDenyKey  = "all_agent_denies:php"
)
