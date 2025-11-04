package database

import (
	"time"

	"jindexer/types"

	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase() (*Database, error) {
	db, err := initDatabase()
	if err != nil {
		return nil, err
	}

	d := Database{
		db: db,
	}

	return &d, nil
}

func (d *Database) SaveBlock(block *types.Block) error {
	return d.db.Create(block).Error
}

func (d *Database) SavePostProof(postProof *types.PostProof) error {
	return d.db.Create(postProof).Error
}

// ListProofsByMerkleAndTimeRange returns all proofs for a given merkle where the referenced block's time
// is between startTime and endTime (inclusive).
func (d *Database) ListProofsByMerkleAndTimeRange(merkle string, startTime, endTime time.Time) ([]types.PostProof, error) {
	var proofs []types.PostProof

	err := d.db.Model(&types.PostProof{}).
		Joins("Block").
		Where("post_proofs.merkle = ?", merkle).
		Where("blocks.time >= ? AND blocks.time <= ?", startTime, endTime).
		Preload("Block").
		Find(&proofs).Error

	return proofs, err
}
