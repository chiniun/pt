package gorm

import (
	"log"
	"os"
	"testing"

	v1 "pt/internal/conf"
	kratoslog "github.com/go-kratos/kratos/v2/log"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Migration struct {
	ID int64 `gorm:"column:id"`
}

func TestMysqlRecoredNotFound(b *testing.T) {
	// 链接数据库
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	bootstrap := &v1.Mysql{Dsn: os.Getenv("TEST_DSN"), Log: &v1.Mysql_Log{IgnoreRecordNotFoundError: false}}
	db, err := NewMySQL(bootstrap, kratoslog.GetLogger())
	if err != nil {
		b.Errorf(err.Error())
	}
	_, err = getBySerializer(db)
	if err != nil {
		b.Errorf(err.Error())
	}

}

func TestMysqlSlowsql(b *testing.T) {
	// 链接数据库
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	bootstrap := &v1.Mysql{Dsn: os.Getenv("TEST_DSN"), Log: &v1.Mysql_Log{ShowSlowSql: true}}
	db, err := NewMySQL(bootstrap, kratoslog.GetLogger())

	_, err = getslowsql(db)
	if err != nil {
		b.Errorf(err.Error())
	}

}

func getBySerializer(db *gorm.DB) (e Migration, err error) {

	e = Migration{}
	result := db.Table("migrations").Where("id = 40").Select("id").First(&e)
	if result.Error != nil {
		return e, result.Error
	}

	return e, nil
}

func getslowsql(db *gorm.DB) (e Migration, err error) {

	e = Migration{}
	result := db.Exec("select sleep(1) from dual")
	if result.Error != nil {
		return e, result.Error
	}

	return e, nil
}
