package src

import (
	"context"
	_ "embed"
	"fmt"
	"log"
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
	contractAddress = common.HexToAddress("0x7A129762332B8f4c6Ed4850c17B218C89e78854d")
)

type contractData struct {
	block       int
	leftMuscle  int64
	rightMuscle int64
	price       float64
	ts          time.Time
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

	bf := &blockFetcher{log: log, client: client, abi: contractAbi}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	latest, err := bf.getLatestBlock(ctx)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(latest)

	data, err := bf.fetchBlockRange(ctx, 13683094, 13683098)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(data)

	return bf, nil
}

func (bf *blockFetcher) MockFetch(startBlock int) (contractData, error) {
	return contractData{
		block:       rand.Int(),
		leftMuscle:  int64(rand.Intn(100)),
		rightMuscle: int64(rand.Intn(100)),
		price:       rand.Float64(),
		ts:          time.Now(),
	}, nil
}

func (bf *blockFetcher) fetch(ch chan contractData, startBlock int) error {
	fmt.Println("fetching from block", startBlock)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	latestBlock, err := bf.getLatestBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block: %w", err)
	}
	cancel()

	fmt.Println("latest block", latestBlock)
	time.Sleep(5 * time.Second)

	// Fetch up to 50 blocks starting from the last one
	for i := startBlock; i < latestBlock; i += 50 {
		from := int64(i)
		to := int64(i + 50)
		if to > int64(latestBlock) {
			to = int64(latestBlock)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cds, err := bf.fetchBlockRange(ctx, from, to)
		if err != nil {
			return fmt.Errorf("failed to fetch block range: %w", err)
		}
		cancel()

		for _, cd := range cds {
			ch <- cd
		}
	}

	return nil
}

func (bf *blockFetcher) fetchBlockRange(ctx context.Context, from, to int64) ([]contractData, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(from),
		ToBlock:   big.NewInt(to),
		Addresses: []common.Address{contractAddress},
	}

	// Fetch logs
	logs, err := bf.client.FilterLogs(ctx, query)
	if err != nil {
		log.Fatalf("Failed to fetch logs: %v", err)
	}

	cds := make([]contractData, 0, len(logs))

	// Decode logs
	for _, vLog := range logs {
		event := struct {
			DeltaX            *big.Int
			DeltaY            *big.Int
			LeftMuscle        *big.Int
			RightMuscle       *big.Int
			PositionTimestamp *big.Int   // timestamp is int?
			PositionPrice     *big.Float // float or int?
		}{}

		if err := bf.abi.UnpackIntoInterface(&event, "WormStateUpdated", vLog.Data); err != nil {
			return nil, fmt.Errorf("Failed to unpack log data: %w", err)
		}

		bigFloatToFloat := func(f *big.Float) float64 {
			fl, _ := f.Float64()
			return fl
		}

		cd := contractData{
			block:       int(vLog.BlockNumber),
			leftMuscle:  event.LeftMuscle.Int64(),
			rightMuscle: event.RightMuscle.Int64(),
			price:       bigFloatToFloat(event.PositionPrice),
			// timestamp TODO
		}

		if cd.leftMuscle == 0 && cd.rightMuscle == 0 {
			bf.log.Warn("zero muscle movements")
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
