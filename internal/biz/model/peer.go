package model

import "time"

type Peer struct {
	ID             uint64    `json:"id"`
	Torrent        uint32    `json:"torrent"`
	PeerID         string    `json:"peer_id"`
	IP             string    `json:"ip"`
	Port           uint16    `json:"port"`
	Uploaded       uint64    `json:"uploaded"`
	Downloaded     uint64    `json:"downloaded"`
	ToGo           uint64    `json:"to_go"`
	Seeder         string    `json:"seeder"`
	Started        time.Time `json:"started"`
	LastAction     time.Time `json:"last_action"`
	PrevAction     time.Time `json:"prev_action"`
	Connectable    string    `json:"connectable"`
	UserID         uint32    `json:"userid"`
	Agent          string    `json:"agent"`
	FinishedAt     uint32    `json:"finishedat"`
	DownloadOffset uint64    `json:"downloadoffset"`
	UploadOffset   uint64    `json:"uploadoffset"`
	Passkey        string    `json:"passkey"`
	IPv4           string    `json:"ipv4"`
	IPv6           string    `json:"ipv6"`
	IsSeedBox      int8      `json:"is_seed_box"`
}

func (*Peer) TableName() string {
	return "peers"
}

type PeerView struct {
	ID                  uint64    `json:"id"`
	Torrent             uint32    `json:"torrent"`
	PeerID              string    `json:"peer_id"`
	IP                  string    `json:"ip"`
	Port                uint16    `json:"port"`
	Uploaded            uint64    `json:"uploaded"`
	Downloaded          uint64    `json:"downloaded"`
	ToGo                uint64    `json:"to_go"`
	Seeder              string    `json:"seeder"`
	Started             time.Time `json:"started"`
	LastAction          time.Time `json:"last_action"`
	LastActionTimeStamp int64     `json:"last_action_unix_timestamp"`
	PrevAction          time.Time `json:"prev_action"`
	Prevts              int64     `json:"prevts"`
	Announcetime        int64     `json:"announcetime"`
	Connectable         string    `json:"connectable"`
	UserID              uint32    `json:"userid"`
	Agent               string    `json:"agent"`
	FinishedAt          uint32    `json:"finishedat"`
	DownloadOffset      uint64    `json:"downloadoffset"`
	UploadOffset        uint64    `json:"uploadoffset"`
	Passkey             string    `json:"passkey"`
	IPv4                string    `json:"ipv4"`
	IPv6                string    `json:"ipv6"`
	IsSeedBox           int8      `json:"is_seed_box"`
}

type PeerBin struct {
	PeerID string `json:"peer_id,omitempty"`
	IP     string `json:"ipv4,omitempty"`
	Port   int32  `json:"port,omitempty"`
}
