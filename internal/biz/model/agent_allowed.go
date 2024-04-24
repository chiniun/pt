package model

type AgentAllowedFamily struct {
	ID              uint    `db:"id" json:"id,omitempty" form:"id"`
	Family          string  `db:"family" json:"family,omitempty" form:"family"`
	StartName       string  `db:"start_name" json:"start_name,omitempty" form:"start_name"`
	PeerIdPattern   string  `db:"peer_id_pattern" json:"peer_id_pattern,omitempty" form:"peer_id_pattern"`
	PeerIdMatchNum  uint8   `db:"peer_id_match_num" json:"peer_id_match_num,omitempty" form:"peer_id_match_num"`
	PeerIdMatchType string  `db:"peer_id_matchtype" json:"peer_id_match_type,omitempty" form:"peer_id_match_type"`
	PeerIdStart     string  `db:"peer_id_start" json:"peer_id_start,omitempty" form:"peer_id_start"`
	AgentPattern    string  `db:"agent_pattern" json:"agent_pattern,omitempty" form:"agent_pattern"`
	AgentMatchNum   uint8   `db:"agent_match_num" json:"agent_match_num,omitempty" form:"agent_match_num"`
	AgentMatchType  string  `db:"agent_matchtype" json:"agent_match_type,omitempty" form:"agent_match_type"`
	AgentStart      string  `db:"agent_start" json:"agent_start,omitempty" form:"agent_start"`
	Exception       string  `db:"exception" json:"exception,omitempty" form:"exception"`
	Allowhttps      string  `db:"allowhttps" json:"allowhttps,omitempty" form:"allowhttps"`
	Comment         *string `db:"comment" json:"comment,omitempty" form:"comment"`
	Hits            uint16  `db:"hits" json:"hits,omitempty" form:"hits"`
}

func (o *AgentAllowedFamily) TableName() string {
	return "agent_allowed_family"
}

type AgentAllowedException struct {
	ID       uint    `db:"id" json:"id,omitempty" form:"id"`
	FamilyID int     `db:"family_id" json:"family_id,omitempty" form:"family_id"`
	Name     string  `db:"name" json:"name,omitempty" form:"name"`
	PeerId   string  `db:"peer_id" json:"peer_id,omitempty" form:"peer_id"`
	Agent    string  `db:"agent" json:"agent,omitempty" form:"agent"`
	Comment  *string `db:"comment" json:"comment,omitempty" form:"comment"`
}

func (o *AgentAllowedException) TableName() string {
	return "agent_allowed_exception"
}
