package constant

type APPROVAL_STATUS int64

const (
	APPROVAL_STATUS_NONE APPROVAL_STATUS = iota
	APPROVAL_STATUS_ALLOW
	APPROVAL_STATUS_DENY
)

const MIN_ANNOUNCE_WAIT_SECOND = 300

const MAX_PEER_NUM_WANT = 50

const MUST_BE_CHEATER_SPEED = 1024 * 1024 * 1024 //1024 MB/s
const MAY_BE_CHEATER_SPEED = 1024 * 1024 * 100   //100 MB/s

const ANNOUNCE_FIRST = 0
const ANNOUNCE_DUAL = 1
const ANNOUNCE_DUPLICATE = 2

// Port Blacklist
var BLACK_PORTS = []int{
	22,                 // SSH Port
	53,                 // DNS queries
	80, 81, 8080, 8081, // Hyper Text Transfer Protocol (HTTP) - port used for web traffic
	411, 412, 413, // 	Direct Connect Hub (unofficial)
	443,        // HTTPS / SSL - encrypted web traffic, also used for VPN tunnels over HTTPS.
	1214,       // Kazaa - peer-to-peer file sharing, some known vulnerabilities, and at least one worm (Benjamin) targeting it.
	3389,       // IANA registered for Microsoft WBT Server, used for Windows Remote Desktop and Remote Assistance connections
	4662,       // eDonkey 2000 P2P file sharing service. http://www.edonkey2000.com/
	6346, 6347, // Gnutella (FrostWire, Limewire, Shareaza, etc.), BearShare file sharing app
	6699, // Port used by p2p software, such as WinMX, Napster.
}

type AnnounceIntervalType int32

const (
	AnnounceInterval      AnnounceIntervalType = 1800
	AnnounceInterTwoAge   AnnounceIntervalType = 7
	AnnounceInterTwo      AnnounceIntervalType = 2700
	AnnounceInterThreeAge AnnounceIntervalType = 30
	AnnounceInterThree    AnnounceIntervalType = 3600
)

const (
	Setting_PaidTorrentEnabled   bool = true
	Setting_IsSeedBoxRuleEnabled      = true
)

type TrackerState int8

const (
	Seeder TrackerState = iota + 1
	Leecher
)

const (
	//欺骗校验
	CheateredSecurity = 2 //cheaterdet_security

	MaximumUploadSpeed     = 8000
	NotSeedBoxMaxSpeedMbps = 100

	TrafficCntPerG = 1073741824
	TrafficCnt10MB = 10485760
	TrafficCntPerMB = 1048576

)
