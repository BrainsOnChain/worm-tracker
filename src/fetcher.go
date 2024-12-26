package src

import (
	"context"
	_ "embed"
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
	contractAddress = common.HexToAddress("0xbc406E6A1592d3b3c2cd20B015da5A7b3983DAC0")

	// Start Block
	initialBlock = 14187700
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

	// Fetch up to 50 blocks starting from the last one
	for i := startBlock; i < latestBlock; i += 50 {
		from := int64(i)
		to := int64(i + 50)
		if to > int64(latestBlock) {
			to = int64(latestBlock)
		}

		cds := bf.fetchBlockRange(context.TODO(), from, to)
		for _, cd := range cds {
			contractDataCh <- cd
		}

		latestBlockCh <- int(to) // Save the last block checked
		time.Sleep(5 * time.Second)
	}

	return nil
}

func (bf *blockFetcher) fetchBlockRange(ctx context.Context, from, to int64) []contractData {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(from),
		ToBlock:   big.NewInt(to),
		Addresses: []common.Address{contractAddress},
	}

	log := bf.log.With(zap.Int64("from", from), zap.Int64("to", to))

	// Fetch logs
	logs, err := bf.client.FilterLogs(ctx, query)
	if err != nil {
		log.Sugar().Warnf("Failed to fetch logs: %v", err)
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

	return cds
}

func (bf *blockFetcher) getLatestBlock(ctx context.Context) (int, error) {
	header, err := bf.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch latest block: %w", err)
	}

	return int(header.Number.Uint64()), nil
}
