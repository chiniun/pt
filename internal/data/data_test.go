package data

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"

	"pt/internal/biz/inter"
	"pt/internal/conf"
)

var (
	// flagconf is the config flag.
	flagconf string
)

var (
	user   inter.UserRepo
	cache  inter.CacheRepo
	ctx    context.Context
	source *Data
)

func TestMain(m *testing.M) {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config_test.yaml")

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	logger := log.NewStdLogger(os.Stdout)

	data, closeFunc, err := NewData(bc.GetData(), logger)
	if err != nil {
		panic(err)
	}
	defer closeFunc()
	source = data
	user = NewUser(data, logger)
	cache = NewCache(data, logger)

	ctx = context.Background()
	os.Exit(m.Run())
}
