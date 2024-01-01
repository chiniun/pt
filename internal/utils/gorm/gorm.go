package gorm

import "gorm.io/gorm"

func RebuildTables(db *gorm.DB, models ...interface{}) (err error) {
	for _, model := range models {
		if dropErr := db.Migrator().DropTable(model); dropErr != nil {
			return dropErr
		}
		if migrateErr := db.AutoMigrate(model); migrateErr != nil {
			return migrateErr
		}
	}
	return
}

func RebuildTablesWithPanic(db *gorm.DB, models ...interface{}) {
	err := RebuildTables(db, models...)
	if err != nil {
		panic(err)
	}
}
