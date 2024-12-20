package src

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

func Run(fetcher *blockFetcher, db *dbManager) error {
	valueCh := make(chan contractData, 10)
	done := make(chan struct{})

	p, err := db.getLatestPosition()
	if err != nil {
		return fmt.Errorf("error getting last block: %w", err)
	}

	dryRun := os.Getenv("DRY_RUN")
	if dryRun == "true" {
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

		go fetcher.fetch(valueCh, p.block)
	}

	for {
		select {
		case contractVal, ok := <-valueCh:
			if !ok {
				return fmt.Errorf("channel closed")
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
		case <-done:
			close(valueCh) // Signal that no more values will be sent
			return nil
		}
	}
}
