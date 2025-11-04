package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"jindexer/types"
)

func initDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("jindexer.db"), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&types.PostProof{},
		&types.Block{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
