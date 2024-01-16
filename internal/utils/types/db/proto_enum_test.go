package db

import (
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type EunmTest struct {
	Id     int64             `gorm:"column:id"`
	Status ProtoEnum[Status] `gorm:"column:status"`
}

func (e EunmTest) TableName() string {
	return "enum_test"
}
func TestEnumdatatype(b *testing.T) {
	dsn := "aa:ss@tcp(fasdfsd:3306)/aa?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		b.Errorf("conn db err:%v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		b.Errorf("get sqldb err:%v", err)
	}
	defer sqlDB.Close()
	db = db.Debug()

	err = inset(db)
	if err != nil {
		b.Errorf("error")
		return
	}

	res, err := get(db)
	if err != nil {
		b.Errorf("get error")
	}

	if res.Status.Data() != Status(1) {
		b.Errorf("status err:%v", res.Status.Data())

	}

	err = update(db, res.Id)
	if err != nil {
		b.Errorf("update err:%v", err)
	}

}

func inset(db *gorm.DB) error {
	// 链接数据库

	result := db.Create(&EunmTest{Status: NewProtoEnum(Status(1))})
	if result.Error != nil {
		return result.Error
	}
	return nil

}

func get(db *gorm.DB) (e *EunmTest, err error) {
	// 链接数据库

	e = &EunmTest{}
	result := db.Model(EunmTest{}).First(e)
	if result.Error != nil {
		return nil, result.Error
	}
	return e, nil

}

func update(db *gorm.DB, id int64) (err error) {
	return db.Model(EunmTest{}).Where("id = ?", id).Updates(EunmTest{Status: NewProtoEnum(Status(1))}).Error

}
