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
