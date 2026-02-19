package database

import (
	"time"

	"github.com/JackalLabs/jindexer/types"

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

// BlockExistsByHeight checks if a block with the given height has been saved before
func (d *Database) BlockExistsByHeight(height int64) (bool, error) {
	var count int64
	err := d.db.Model(&types.Block{}).Where("height = ?", height).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetMostRecentBlockHeight returns the height of the most recently saved block.
// Returns 0 and an error if no blocks are found or if there's a database error.
func (d *Database) GetMostRecentBlockHeight() (int64, error) {
	var block types.Block
	err := d.db.Model(&types.Block{}).
		Order("height DESC").
		First(&block).Error
	if err != nil {
		return 0, err
	}
	return block.Height, nil
}

func (d *Database) SavePostProof(postProof *types.PostProof) error {
	return d.db.Create(postProof).Error
}

// ListProofsByMerkleAndTimeRange returns all proofs for a given merkle where the referenced block's time
// is between startTime and endTime (inclusive), ordered by block date (most recent first).
func (d *Database) ListProofsByMerkleAndTimeRange(merkle string, startTime, endTime time.Time) ([]types.PostProof, error) {
	var proofs []types.PostProof

	err := d.db.Model(&types.PostProof{}).
		Joins("INNER JOIN blocks ON post_proofs.block_id = blocks.id").
		Where("post_proofs.merkle = ?", merkle).
		Where("blocks.time >= ? AND blocks.time <= ?", startTime, endTime).
		Order("blocks.time DESC").
		Preload("Block").
		Find(&proofs).Error

	return proofs, err
}

// ListRecentProofs returns the most recent proofs ordered by block date (most recent first), limited to the specified count.
func (d *Database) ListRecentProofs(limit int) ([]types.PostProof, error) {
	var proofs []types.PostProof

	err := d.db.Model(&types.PostProof{}).
		Joins("INNER JOIN blocks ON post_proofs.block_id = blocks.id").
		Order("blocks.time DESC").
		Limit(limit).
		Preload("Block").
		Find(&proofs).Error

	return proofs, err
}

// ListProofsByID returns proofs ordered by ID (most recent first), limited to the specified count.
func (d *Database) ListProofsByID(limit int) ([]types.PostProof, error) {
	var proofs []types.PostProof

	err := d.db.Model(&types.PostProof{}).
		Order("id DESC").
		Limit(limit).
		Preload("Block").
		Find(&proofs).Error

	return proofs, err
}

// MerkleLastProof holds the most recent proof timestamp for a merkle.
type MerkleLastProof struct {
	Merkle        string
	LastProofTime time.Time
}

// GetMerkleLastProofTimes returns the most recent block time per merkle using
// a SQL aggregate instead of loading individual rows.
func (d *Database) GetMerkleLastProofTimes() ([]MerkleLastProof, error) {
	var results []MerkleLastProof

	err := d.db.Model(&types.PostProof{}).
		Select("post_proofs.merkle, MAX(blocks.time) as last_proof_time").
		Joins("INNER JOIN blocks ON post_proofs.block_id = blocks.id").
		Group("post_proofs.merkle").
		Scan(&results).Error

	return results, err
}

// GetTotalProofCount returns the total number of proofs in the database.
func (d *Database) GetTotalProofCount() (int64, error) {
	var count int64
	err := d.db.Model(&types.PostProof{}).Count(&count).Error
	return count, err
}
