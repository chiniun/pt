package data

import (
	"context"
	"fmt"
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
