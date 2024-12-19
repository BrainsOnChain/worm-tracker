package src

import (
	"fmt"
)

func Run(fetcher *blockFetcher, db *dbManager) error {
	valueCh := make(chan contractData, 10)
	done := make(chan struct{})

	p, err := db.getLatestPosition()
	if err != nil {
		return fmt.Errorf("error getting last block: %w", err)
	}

	go fetcher.fetch(valueCh, p.block)

	for {
		select {
		case contractVal, ok := <-valueCh:
			if !ok {
				return fmt.Errorf("channel closed")
			}

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
