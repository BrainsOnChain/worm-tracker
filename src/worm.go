package src

import (
	"fmt"
	"time"
)

func Run(fetcher *blockFetcher, db *dbManager) error {

	// 1. Fetch the latest position information including the last known block
	// 2. Fetch the blocks from last block to latest
	// 3. For each block if there is left, right muscle values > 0 update the position and save that position

	var p position
	for {
		cd, err := fetcher.MockFetch(0)
		if err != nil {
			return fmt.Errorf("error fetchig mock data: %w", err)
		}

		p = updatePosition(cd, p)
		if err := db.savePosition(p); err != nil {
			return fmt.Errorf("error saving position: %w", err)
		}

		time.Sleep(5 * time.Second)
	}
}
