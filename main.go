package main

import (
	"context"
	"os"
	"strconv"

	"github.com/JackalLabs/jindexer/database"
	"github.com/JackalLabs/jindexer/indexer"
	"github.com/JackalLabs/jindexer/utils"
	"github.com/cosmos/cosmos-sdk/client"
	canine "github.com/jackalLabs/canine-chain/v5/app"
	"github.com/rs/zerolog/log"
)

func main() {
	utils.InitLogger("Starting JIndexer")

	// Get RPC and gRPC endpoints from environment variables
	rpcEndpoint := os.Getenv("JACKAL_RPC_URL")
	if rpcEndpoint == "" {
		rpcEndpoint = "https://jackal-rpc.polkachu.com:443"
	}

	grpcEndpoint := os.Getenv("JACKAL_GRPC_URL")
	if grpcEndpoint == "" {
		grpcEndpoint = "jackal-grpc.polkachu.com:17590"
	}

	// Get start height from environment variable
	startHeightStr := os.Getenv("JINDEXER_START_HEIGHT")
	var startHeight int64
	if startHeightStr == "" {
		startHeight = 0 // Default to 0, which means current block height
	} else {
		var err error
		startHeight, err = strconv.ParseInt(startHeightStr, 10, 64)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse JINDEXER_START_HEIGHT")
		}
	}

	encodingCfg := canine.MakeEncodingConfig()

	d, err := database.NewDatabase()
	if err != nil {
		panic(err)
	}

	// If startHeight is 0, try to get the most recent block from the database first,
	// then fall back to current block height from RPC if there's an error
	if startHeight == 0 {
		mostRecentHeight, err := d.GetMostRecentBlockHeight()
		if err == nil {
			// Start from the next block after the most recent one
			startHeight = mostRecentHeight + 1
			log.Info().Int64("start_height", startHeight).Int64("last_indexed_height", mostRecentHeight).Msg("Starting after most recently saved block")
		} else {
			// If there's an error (e.g., no blocks in database), fall back to current block height from RPC
			log.Warn().Err(err).Msg("Failed to get most recent block from database, falling back to current block height from RPC")

			rpcClient, err := client.NewClientFromNode(rpcEndpoint)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to create RPC client to get current block height")
			}

			ctx := context.Background()
			abciInfo, err := rpcClient.ABCIInfo(ctx)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get current block height from RPC")
			}

			startHeight = abciInfo.Response.LastBlockHeight
			log.Info().Int64("start_height", startHeight).Msg("Starting from current block height")
		}
	}

	i, err := indexer.NewIndexer(rpcEndpoint, grpcEndpoint, encodingCfg, d, startHeight, 0)
	if err != nil {
		panic(err)
	}

	i.Start()
}
