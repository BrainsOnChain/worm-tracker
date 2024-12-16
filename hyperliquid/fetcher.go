package hyperliquid

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

//go:embed abi.json
var abiStr string

func Fetch() {
	// Connect to Hyperliquid or any Ethereum-compatible blockchain
	client, err := ethclient.Dial("https://api.hyperliquid-testnet.xyz/evm")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// Contract address
	contractAddress := common.HexToAddress("0x7A129762332B8f4c6Ed4850c17B218C89e78854d")

	// Filter query
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(13284970), // Specify block range
		ToBlock:   big.NewInt(13284974),
		Addresses: []common.Address{contractAddress},
	}

	// Fetch logs
	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatalf("Failed to fetch logs: %v", err)
	}

	// ABI parsing
	contractAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	// Decode logs
	for _, vLog := range logs {
		event := struct {
			DeltaX            *big.Int
			DeltaY            *big.Int
			LeftMuscle        *big.Int
			RightMuscle       *big.Int
			PositionTimestamp *big.Int
			PositionPrice     *big.Int
		}{}

		err := contractAbi.UnpackIntoInterface(&event, "WormStateUpdated", vLog.Data)
		if err != nil {
			log.Fatalf("Failed to unpack log data: %v", err)
		}

		fmt.Printf("Event Data: %v %+v\n", vLog.BlockNumber, event)
	}
}

func FollowChainWithPolling() {
	// Connect to the Ethereum-compatible blockchain
	client, err := ethclient.Dial("https://api.hyperliquid-testnet.xyz/evm")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// Contract address
	contractAddress := common.HexToAddress("0x7A129762332B8f4c6Ed4850c17B218C89e78854d")

	// Parse contract ABI
	contractAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	// Start block
	var lastBlock uint64 = 13554153 // Start from a known block
	fmt.Println("Starting block:", lastBlock)

	// Polling loop
	for {
		// Get the latest block number
		header, err := client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			log.Printf("Failed to fetch latest block: %v", err)
			time.Sleep(5 * time.Second) // Retry after delay
			continue
		}
		latestBlock := header.Number.Uint64()

		// Check for new blocks
		if latestBlock > lastBlock {
			fmt.Printf("Fetching logs from blocks %d to %d...\n", lastBlock+1, latestBlock)

			// Define filter query
			query := ethereum.FilterQuery{
				FromBlock: big.NewInt(int64(lastBlock + 1)),
				ToBlock:   big.NewInt(int64(latestBlock)),
				Addresses: []common.Address{contractAddress},
			}

			// Fetch logs
			logs, err := client.FilterLogs(context.Background(), query)
			if err != nil {
				log.Printf("Failed to fetch logs: %v", err)
				time.Sleep(5 * time.Second) // Retry after delay
				continue
			}

			fmt.Println("logs:", len(logs))
			// Process logs
			for _, vLog := range logs {
				event := struct {
					DeltaX            *big.Int
					DeltaY            *big.Int
					LeftMuscle        *big.Int
					RightMuscle       *big.Int
					PositionTimestamp *big.Int
					PositionPrice     *big.Int
				}{}

				err := contractAbi.UnpackIntoInterface(&event, "WormStateUpdated", vLog.Data)
				if err != nil {
					log.Printf("Failed to unpack log data: %v", err)
					continue
				}

				fmt.Printf("New Event - Block: %d, Data: %+v\n", vLog.BlockNumber, event)
			}

			// Update last processed block
			lastBlock = latestBlock
		}

		// Wait for a short interval before polling again
		time.Sleep(5 * time.Second)
	}
}
