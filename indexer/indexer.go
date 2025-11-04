package indexer

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/JackalLabs/jindexer/database"
	types2 "github.com/JackalLabs/jindexer/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/jackalLabs/canine-chain/v5/app/params"
	"github.com/jackalLabs/canine-chain/v5/x/storage/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/tendermint/rpc/client/http"
	"google.golang.org/grpc"
)

type Indexer struct {
	running       bool
	startHeight   int64
	endHeight     int64
	currentHeight int64
	grpcClient    *grpc.ClientConn
	rpcClient     *http.HTTP
	codec         params.EncodingConfig
	database      *database.Database
}

func NewIndexer(rpcEndpoint string, grpcEndpoint string, codec params.EncodingConfig, db *database.Database, startHeight int64, endHeight int64) (*Indexer, error) {
	rpcClient, err := client.NewClientFromNode(rpcEndpoint)
	if err != nil {
		return nil, err
	}

	grpcClient, err := CreateGrpcConnection(grpcEndpoint)
	if err != nil {
		return nil, err
	}

	i := Indexer{
		running:       false,
		startHeight:   startHeight,
		endHeight:     endHeight,
		currentHeight: startHeight,
		grpcClient:    grpcClient,
		rpcClient:     rpcClient,
		codec:         codec,
		database:      db,
	}
	return &i, nil
}

func (i *Indexer) Start() {
	i.running = true
	ctx := context.Background()
	for i.running {
		if i.currentHeight >= i.endHeight && i.endHeight > 0 { // stop when end height is reached if end height is not 0
			i.running = false
			return
		}

		i.indexBlock(ctx, i.currentHeight)
		i.currentHeight++
	}
}

func (i *Indexer) indexBlock(ctx context.Context, height int64) {
	log.Info().Int64("height", height).Msg("Indexing block...")

	var networkHeight int64 = 0

	for networkHeight < height || networkHeight == 0 {
		if networkHeight < height && networkHeight > 0 {
			log.Info().Int64("current_height", height).Int64("network_height", networkHeight).Msg("network is behind us, waiting for more blocks")
			time.Sleep(time.Second * 6)
		}
		abciInfo, err := i.rpcClient.ABCIInfo(ctx)
		if err != nil {
			log.Err(err).Msg("failed to get abci info")
			return
		}

		networkHeight = abciInfo.Response.LastBlockHeight
	}

	blockInfo, err := i.rpcClient.Block(ctx, &height)
	if err != nil {
		log.Err(err).Msg("failed to get block info")
		return
	}

	block := blockInfo.Block

	b := types2.Block{
		Time:   block.Time,
		Height: height,
	}
	err = i.database.SaveBlock(&b)
	if err != nil {
		log.Err(err).Msg("failed to save block info")
		return
	}

	txs := block.Txs
	log.Info().Int("TX_Count", len(txs)).Msg("Indexed block.")

	for _, txBytes := range txs {
		txHash := hex.EncodeToString(txBytes.Hash())

		// Decode the transaction bytes into a Tx
		tx, err := i.codec.TxConfig.TxDecoder()(txBytes)
		if err != nil {
			log.Err(err).Str("tx", txHash).Msg("failed to decode TX")
			continue
		}

		// Extract messages from the transaction
		msgs := tx.GetMsgs()
		for _, msg := range msgs {
			err := i.processMessage(msg, b)
			if err != nil {
				log.Err(err).Msg("could not process message")
				continue
			}
		}

		log.Info().Str("tx", txHash).Msg("Tx parsed")
	}
}

func (i *Indexer) processMessage(msg sdk.Msg, block types2.Block) error {
	// Get the type URL from the message by packing it into an Any
	msgAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		log.Err(err).Msg("failed to pack message into Any")
		return err
	}
	messageType := msgAny.TypeUrl

	log.Info().Str("message_type", messageType).Msg("processing message")

	switch messageType {
	case "/canine_chain.storage.MsgPostProof":
		err = i.processPostProof(msg, block)
	default:
		log.Warn().Str("message_type_url", messageType).Msg("could not process message")
		return nil
	}

	return err
}

func (i *Indexer) processPostProof(msg sdk.Msg, block types2.Block) error {
	// Cast the message to the specific type
	msgPostProof, ok := msg.(*types.MsgPostProof)
	if !ok {
		return nil
	}

	// Process the message
	log.Info().Msg("processing MsgPostProof")
	_ = msgPostProof // Use msgPostProof as needed

	merkle := hex.EncodeToString(msgPostProof.Merkle)
	prover := msgPostProof.Creator

	postProof := types2.PostProof{
		Merkle: merkle,
		Prover: prover,
		Block:  block,
	}

	err := i.database.SavePostProof(&postProof)
	if err != nil {
		return err
	}

	return nil
}
