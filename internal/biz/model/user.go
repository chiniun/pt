package model

import "time"

type User struct {
	Id                  int64      `json:"id"`
	Username            string     `json:"username"`
	Passhash            string     `json:"passhash"`
	Secret              string     `json:"secret"`
	Email               string     `json:"email"`
	Status              string     `json:"status"`
	Added               *time.Time `json:"added"`
	LastLogin           *time.Time `json:"last_login"`
	LastAccess          *time.Time `json:"last_access"`
	LastHome            *time.Time `json:"last_home"`
	LastOffer           *time.Time `json:"last_offer"`
	ForumAccess         *time.Time `json:"forum_access"`
	LastStaffMsg        *time.Time `json:"last_staffmsg"`
	LastPM              *time.Time `json:"last_pm"`
	LastComment         *time.Time `json:"last_comment"`
	LastPost            *time.Time `json:"last_post"`
	LastBrowse          int64      `json:"last_browse"`
	LastMusic           int64      `json:"last_music"`
	LastCatchup         int64      `json:"last_catchup"`
	EditSecret          string     `json:"editsecret"`
	Privacy             string     `json:"privacy"`
	Stylesheet          int64      `json:"stylesheet"`
	Caticon             int64      `json:"caticon"`
	FontSize            string     `json:"fontsize"`
	Info                string     `json:"info"`
	AcceptPMS           string     `json:"acceptpms"`
	CommentPM           string     `json:"commentpm"`
	IP                  string     `json:"ip"`
	Class               int64      `json:"class"`
	MaxClassOnce        int64      `json:"max_class_once"`
	Avatar              string     `json:"avatar"`
	Uploaded            int64      `json:"uploaded"`
	Downloaded          int64      `json:"downloaded"`
	Seedtime            int64      `json:"seedtime"`
	Leechtime           int64      `json:"leechtime"`
	Title               string     `json:"title"`
	Country             int64      `json:"country"`
	Notifs              string     `json:"notifs"`
	ModComment          string     `json:"modcomment"`
	Enabled             string     `json:"enabled"`
	Avatars             string     `json:"avatars"`
	Donor               string     `json:"donor"`
	Donated             float64    `json:"donated"`
	DonatedCNY          float64    `json:"donated_cny"`
	DonorUntil          *time.Time `json:"donoruntil"`
	Warned              string     `json:"warned"`
	WarnedUntil         *time.Time `json:"warneduntil"`
	NoAd                string     `json:"noad"`
	NoAdUntil           *time.Time `json:"noaduntil"`
	TorrentsPerPage     int64      `json:"torrentsperpage"`
	TopicsPerPage       int64      `json:"topicsperpage"`
	PostsPerPage        int64      `json:"postsperpage"`
	ClickTopic          string     `json:"clicktopic"`
	DeletePMS           string     `json:"deletepms"`
	SavePMS             string     `json:"savepms"`
	ShowHot             string     `json:"showhot"`
	ShowClassic         string     `json:"showclassic"`
	Support             string     `json:"support"`
	Picker              string     `json:"picker"`
	StaffFor            string     `json:"stafffor"`
	SupportFor          string     `json:"supportfor"`
	PickFor             string     `json:"pickfor"`
	SupportLang         string     `json:"supportlang"`
	Passkey             string     `json:"passkey"`
	PromotionLink       string     `json:"promotion_link"`
	UploadPos           string     `json:"uploadpos"`
	ForumPost           string     `json:"forumpost"`
	DownloadPos         string     `json:"downloadpos"`
	ClientSelect        int64      `json:"clientselect"`
	Signatures          string     `json:"signatures"`
	Signature           string     `json:"signature"`
	Lang                uint16     `json:"lang"`
	Cheat               int16      `json:"cheat"`
	Download            int64      `json:"download"`
	Upload              int64      `json:"upload"`
	ISP                 int64      `json:"isp"`
	Invites             int64      `json:"invites"`
	InvitedBy           int64      `json:"invited_by"`
	Gender              string     `json:"gender"`
	VIPAdded            string     `json:"vip_added"`
	VIPUntil            *time.Time `json:"vip_until"`
	SeedBonus           float64    `json:"seedbonus"`
	Charity             float64    `json:"charity"`
	BonusComment        string     `json:"bonuscomment"`
	Parked              string     `json:"parked"`
	LeechWarn           string     `json:"leechwarn"`
	LeechWarnUntil      *time.Time `json:"leechwarnuntil"`
	LastWarned          *time.Time `json:"lastwarned"`
	TimesWarned         int64      `json:"timeswarned"`
	WarnedBy            int64      `json:"warnedby"`
	SBNum               int64      `json:"sbnum"`
	SBRefresh           int64      `json:"sbrefresh"`
	HideHB              string     `json:"hidehb"`
	ShowIMDB            string     `json:"showimdb"`
	ShowDescription     string     `json:"showdescription"`
	ShowComment         string     `json:"showcomment"`
	ShowClientError     string     `json:"showclienterror"`
	ShowDLNotice        int64      `json:"showdlnotice"`
	Tooltip             string     `json:"tooltip"`
	ShowNFO             string     `json:"shownfo"`
	TimeType            string     `json:"timetype"`
	AppendSticky        string     `json:"appendsticky"`
	AppendNew           string     `json:"appendnew"`
	AppendPromotion     string     `json:"appendpromotion"`
	AppendPicked        string     `json:"appendpicked"`
	DLIcon              string     `json:"dlicon"`
	BMIcon              string     `json:"bmicon"`
	ShowSmallDescr      string     `json:"showsmalldescr"`
	ShowComNum          string     `json:"showcomnum"`
	ShowLastCom         string     `json:"showlastcom"`
	ShowLastPost        string     `json:"showlastpost"`
	PMNum               uint8      `json:"pmnum"`
	School              uint16     `json:"school"`
	ShowFB              string     `json:"showfb"`
	Page                string     `json:"page"`
	TwoStepSecret       string     `json:"two_step_secret"`
	SeedPoints          float64    `json:"seed_points"`
	SeedPointsPerHour   float64    `json:"seed_points_per_hour"`
	AttendanceCard      int64      `json:"attendance_card"`
	OfferAllowedCount   int64      `json:"offer_allowed_count"`
	SeedPointsUpdatedAt *time.Time `json:"seed_points_updated_at"`
	SeedTimeUpdatedAt   *time.Time `json:"seed_time_updated_at"`
}

func (o *User) TableName() string {
	return "users"
}

type BonusLog struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	BusinessType  int64     `gorm:"column:business_type;not null;default:0"`
	UID           int64     `gorm:"column:uid;not null"`
	OldTotalValue float64   `gorm:"column:old_total_value;not null;type:decimal(20,1)"`
	Value         float64   `gorm:"column:value;not null;type:decimal(20,1)"`
	NewTotalValue float64   `gorm:"column:new_total_value;not null;type:decimal(20,1)"`
	Comment       string    `gorm:"column:comment"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:current_timestamp"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;default:current_timestamp;onUpdate:current_timestamp"`
}

func (o *BonusLog) TableName() string {
	return "bonus_logs"
}
