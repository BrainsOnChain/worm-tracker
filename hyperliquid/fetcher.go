package hyperliquid

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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

	// Event signature hash (Keccak-256 hash of the event signature)
	eventSig := "WormStateUpdatedByUser(int256,int256,uint256,uint256,uint256,address)"
	eventHash := common.HexToHash("0x" + fmt.Sprintf("%x", crypto.Keccak256([]byte(eventSig))))

	// Filter query
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(13284970), // Specify block range
		ToBlock:   big.NewInt(13284974),
		Addresses: []common.Address{contractAddress},
		Topics:    [][]common.Hash{{eventHash}},
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
			DeltaX         *big.Int
			DeltaY         *big.Int
			LeftMuscle     *big.Int
			RightMuscle    *big.Int
			Timestamp      *big.Int
			TriggeringUser common.Address
		}{}

		err := contractAbi.UnpackIntoInterface(&event, "WormStateUpdatedByUser", vLog.Data)
		if err != nil {
			log.Fatalf("Failed to unpack log data: %v", err)
		}

		fmt.Printf("Event Data: %+v\n", event)
	}
}
