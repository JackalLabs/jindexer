package types

import (
	"time"

	"gorm.io/gorm"
)

type Block struct {
	gorm.Model

	Height int64     `json:"height" gorm:"uniqueIndex"`
	Time   time.Time `json:"time"`
}

type PostProof struct {
	gorm.Model

	Merkle string `json:"merkle" gorm:"index"`
	Prover string `json:"prover" gorm:"index"`

	Block   Block `json:"block"`
	BlockId uint  `json:"blockId"`
}
