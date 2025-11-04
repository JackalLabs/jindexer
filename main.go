package main

import (
	"github.com/JackalLabs/jindexer/database"
	"github.com/JackalLabs/jindexer/indexer"
	"github.com/JackalLabs/jindexer/utils"
	canine "github.com/jackalLabs/canine-chain/v5/app"
)

func main() {
	utils.InitLogger("Starting JIndexer")

	var startHeight int64 = 15233690

	encodingCfg := canine.MakeEncodingConfig()

	d, err := database.NewDatabase()
	if err != nil {
		panic(err)
	}

	i, err := indexer.NewIndexer("https://jackal-rpc.polkachu.com:443", "jackal-grpc.polkachu.com:17590", encodingCfg, d, startHeight, 0)
	if err != nil {
		panic(err)
	}

	i.Start()
}
