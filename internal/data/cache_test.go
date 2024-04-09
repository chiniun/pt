package data

import (
	"context"
	"encoding/json"
	"fmt"
	"pt/internal/biz"
	"testing"
)

func TestCache_HGet(t *testing.T) {

	data, err := cache.HGet(context.TODO(), "key", "xxx")
	if err != nil {
		fmt.Println(fmt.Sprintf("%#+v", err))
	}
	t.Log(data)
}

func TestCache_HSet(t *testing.T) {

	data, err := cache.HSet(context.TODO(), "key", "xx", true)
	if err != nil {
		fmt.Println(fmt.Sprintf("%#+v", err))
	}
	t.Log(data)
}

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

func Test_Json(t *testing.T) {
	x := biz.AnnounceRequest{
		InfoHash:      "xxxx",
		PeerID:        "xxx",
		IP:            "xx",
		Port:          57643,
		Uploaded:      1000,
		Downloaded:    1000,
		Left:          9000,
		Numwant:       30,
		Key:           "xx",
		Compact:       false,
		SupportCrypto: false,
		Event:         "do",

		Passkey:  "",
		Authkey:  "",
		RawQuery: "",
	}
	d, _ := json.Marshal(x)
	fmt.Println(string(d))
}
