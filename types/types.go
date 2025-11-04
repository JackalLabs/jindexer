package types

import "time"
import "gorm.io/gorm"

type Block struct {
	gorm.Model

	Height int64     `json:"height"`
	Time   time.Time `json:"time"`
}

type PostProof struct {
	gorm.Model

	Merkle string `json:"merkle"`
	Prover string `json:"prover"`

	Block   Block `json:"block"`
	BlockId uint  `json:"blockId"`
}
