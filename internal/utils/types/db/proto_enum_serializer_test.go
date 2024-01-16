package db

import (
	"testing"

	
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Status int64

type EunmSerializerTest struct {
	Id     int64  `gorm:"column:id"`
	Status Status `gorm:"column:status;serializer:enum"`
}

func (e EunmSerializerTest) TableName() string {
	return "enum_test"
}
func TestEnumserializer(b *testing.T) {
	// 链接数据库
	dsn := "aa:ss@tcp(fasdfsd:3306)/aa?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		b.Errorf("%v", err)
	}

	db = db.Debug()

	sqlDB, err := db.DB()
	if err != nil {
		b.Errorf("%v", err)
	}
	defer sqlDB.Close()

	err = insetBySerializer(db)
	if err != nil {
		b.Errorf("error")
		return
	}

	e, err := getBySerializer(db)
	if err != nil {
		b.Errorf("get error")
	}

	if e.Status != Status(1) {
		b.Errorf("insert data err")

	}

	// err = updateBySerializer(db, e.Id)
	// if err != nil {
	// 	b.Errorf("update err:%v", err)
	// }

	// e, err = getBySerializer(db)
	// if err != nil {
	// 	b.Errorf("get error")
	// }

	// if e.Status != migration.Status_STATUS_MIGRATING_ROLLBACKING {
	// 	b.Errorf("update data err")

	// }

}
func insetBySerializer(db *gorm.DB) error {

	result := db.Create(&EunmSerializerTest{Status: Status(1)})
	if result.Error != nil {
		return result.Error
	}
	return nil

}

func getBySerializer(db *gorm.DB) (e EunmSerializerTest, err error) {

	e = EunmSerializerTest{}
	result := db.Model(EunmSerializerTest{}).Order("id desc").First(&e)
	if result.Error != nil {
		return e, result.Error
	}

	return e, nil
}

func updateBySerializer(db *gorm.DB, id int64) (err error) {
	return db.Model(EunmSerializerTest{}).Where("id = ?", id).
		Updates(EunmSerializerTest{Status: Status(1)}).Error
}
