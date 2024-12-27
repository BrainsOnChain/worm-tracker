package src

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

const (
	hypeAPI = "https://api.hyperliquid-testnet.xyz/evm"
)

var (
	//go:embed abi.json
	abiStr string

	// Contract address
	contractAddress = common.HexToAddress("0x385B69Ef54332E6D3f00Ecf3384F890183e511F8")

	// Start Block
	initialBlock = 14419337
)

type contractData struct {
	transactionHash string
	block           int
	leftMuscle      int64
	rightMuscle     int64
	price           float64
	ts              time.Time
}

type blockFetcher struct {
	log    *zap.Logger
	client *ethclient.Client
	abi    abi.ABI
}

func NewBlockFetcher(log *zap.Logger) (*blockFetcher, error) {
	// Connect to Hyperliquid or any Ethereum-compatible blockchain
	client, err := ethclient.Dial(hypeAPI)
	if err != nil {
		return nil, fmt.Errorf("error connecting to hype client: %w", err)
	}

	// ABI parsing
	contractAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract ABI: %w", err)
	}

	return &blockFetcher{log: log, client: client, abi: contractAbi}, nil
}

func (bf *blockFetcher) mockFetch() (contractData, error) {
	return contractData{
		block:       rand.Int(),
		leftMuscle:  int64(rand.Intn(100)),
		rightMuscle: int64(rand.Intn(100)),
		price:       rand.Float64(),
		ts:          time.Now(),
	}, nil
}

// fetch fetches the contract data from the blockchain and sends it to the
// contractData channel. It also sends the latest block checked to the
// latestBlock channel. It does so in batches of 50 blocks. However, if it
// encounters an invalid block range, it will switch to single block fetching to
// find the problematic block.
func (bf *blockFetcher) fetch(contractDataCh chan contractData, latestBlockCh chan int, startBlock int) error {
	if startBlock == 0 {
		startBlock = initialBlock
	}

	latestBlock, err := bf.getLatestBlock(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get latest block: %w", err)
	}

	bf.log.Info(
		"fetching blocks",
		zap.Int("start", startBlock),
		zap.Int("latest", latestBlock),
		zap.Int("to_query", latestBlock-startBlock),
	)

	// Start with batch size of 50
	i := startBlock
	batchSize := 50

	for i < latestBlock {
		from, to := i, i+batchSize
		if to > latestBlock {
			to = latestBlock
		}

		cds, err := bf.fetchBlockRange(context.Background(), int64(from), int64(to))
		if err != nil {
			if errors.Is(err, errInvalidBlockRange) {
				// If we hit an invalid block range and we're not already at
				// batch size 1
				if batchSize > 1 {
					bf.log.Info(
						"switching to single block fetching to find problematic block",
						zap.Int("from_block", i),
					)
					batchSize = 1
					continue // Retry fetching with batch size 1
				} else {
					// We found the problematic block, log it and skip it
					bf.log.Warn("found invalid block, skipping", zap.Int("block", i))
					i++
					continue // Skip to next block
				}
			}
			return fmt.Errorf("failed to fetch block range: %w", err)
		}

		// If we successfully fetched with batch size 1, try increasing batch
		// size again
		if batchSize == 1 {
			batchSize = 50
			bf.log.Info("resuming batch fetching", zap.Int("at_block", i))
		}

		for _, cd := range cds {
			contractDataCh <- cd
		}

		latestBlockCh <- to // Save the last block checked
		i += batchSize
		time.Sleep(1 * time.Second)
	}

	return nil
}

var errInvalidBlockRange = errors.New("invalid block range")

func (bf *blockFetcher) fetchBlockRange(ctx context.Context, from, to int64) ([]contractData, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(from),
		ToBlock:   big.NewInt(to),
		Addresses: []common.Address{contractAddress},
	}

	log := bf.log.With(zap.Int64("from", from), zap.Int64("to", to))

	// Fetch logs
	logs, err := bf.client.FilterLogs(ctx, query)
	if err != nil {
		if strings.Contains(err.Error(), "invalid block range") {
			return nil, errInvalidBlockRange
		}
		return nil, err // returning here will cause the fetch to return
	}
	log.Info("fetching block range", zap.Int("logs", len(logs)))

	cds := make([]contractData, 0, len(logs))

	// Decode logs
	for _, vLog := range logs {
		event := struct {
			DeltaX            *big.Int
			DeltaY            *big.Int
			LeftMuscle        *big.Int
			RightMuscle       *big.Int
			PositionTimestamp *big.Int // timestamp is int?
			PositionPrice     *big.Int // float or int?
		}{}

		if err := bf.abi.UnpackIntoInterface(&event, "WormStateUpdated", vLog.Data); err != nil {
			log.Sugar().Warnf("failed to unpack log data: %w", err)
			continue
		}

		cd := contractData{
			transactionHash: vLog.TxHash.String(),
			block:           int(vLog.BlockNumber),
			leftMuscle:      event.LeftMuscle.Int64(),
			rightMuscle:     event.RightMuscle.Int64(),
			price:           float64(event.PositionPrice.Int64()) / 10000000,
			ts:              time.Unix(event.PositionTimestamp.Int64(), 0), // Set ts by converting the UNIX timestamp
		}

		if cd.leftMuscle == 0 && cd.rightMuscle == 0 {
			log.Info("zero muscle movements, ignoring", zap.Int("block", cd.block))
			continue
		}

		cds = append(cds, cd)
	}

	return cds, nil
}

func (bf *blockFetcher) getLatestBlock(ctx context.Context) (int, error) {
	header, err := bf.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch latest block: %w", err)
	}

	return int(header.Number.Uint64()), nil
}
