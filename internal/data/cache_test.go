package data

import (
	"context"
	"fmt"
	"testing"
)

func TestCache_Get(t *testing.T) {

	data, err := cache.HGet(context.TODO(), "key", "xx")
	if err != nil {
		fmt.Println(fmt.Sprintf("%#+v", err))
	}
	t.Log(data)
}
