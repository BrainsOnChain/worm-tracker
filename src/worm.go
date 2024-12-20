package src

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

func Run(fetcher *blockFetcher, db *dbManager) error {
	valueCh := make(chan contractData, 10)
	blockCh := make(chan int)

	latestBlock, err := db.getLatestBlockChecked()
	if err != nil {
		return fmt.Errorf("error getting last block: %w", err)
	}

	p, err := db.getLatestPosition()
	if err != nil {
		return fmt.Errorf("error getting latest position: %w", err)
	}

	if dryRun := os.Getenv("DRY_RUN"); dryRun == "true" {
		zap.L().Info("starting fetcher in dry-run mode")

		go func() {
			for {
				cd, err := fetcher.mockFetch()
				if err != nil {
					fmt.Println(err)
				}
				valueCh <- cd
				time.Sleep(5 * time.Second)
			}
		}()

	} else {
		zap.L().Info("starting fetcher in live mode")

		// run the fetcher in a goroutine but if it returns nil start it again
		// after a 1 minute sleep this is to handle the case where the latest
		// checked block is the current block
		go func() {
			for {
				if err := fetcher.fetch(valueCh, blockCh, latestBlock); err == nil {
					zap.L().Info("fetcher returned, sleeping for 1 minute")
					time.Sleep(1 * time.Minute)
				}
			}
		}()
	}

	for {
		select {
		case block, ok := <-blockCh:
			if !ok {
				return fmt.Errorf("block channel closed")
			}

			zap.L().Info("new block tracked", zap.Int("block", block))

			if err := db.saveBlockChecked(block); err != nil {
				return fmt.Errorf("error saving block: %w", err)
			}
		case contractVal, ok := <-valueCh:
			if !ok {
				return fmt.Errorf("contract data channel closed")
			}

			zap.L().Info(
				"received contract data",
				zap.Int("block", contractVal.block),
				zap.Int64("left_muscle", contractVal.leftMuscle),
				zap.Int64("right_muscle", contractVal.rightMuscle),
				zap.Float64("price", contractVal.price),
				zap.Time("ts", contractVal.ts),
			)

			p = updatePosition(contractVal, p)
			if err := db.savePosition(p); err != nil {
				return fmt.Errorf("error saving position: %w", err)
			}
		}
	}
}
