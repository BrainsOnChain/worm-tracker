package src

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type cache struct {
	log    *zap.Logger
	db     *dbManager
	mutex  *sync.Mutex
	ticker *time.Ticker
	data   []position
}

func NewCache(log *zap.Logger, db *dbManager) *cache {
	c := &cache{
		log:    log,
		db:     db,
		mutex:  &sync.Mutex{},
		ticker: time.NewTicker(2 * time.Minute),
	}
	c.refresh() // initial refresh

	return c
}

// Run the cache obtains the last 1000 positios from the database using a ticker
// at in interval of once very 5 minutes and updates the cache with the new data
func (c *cache) Run() {
	for range c.ticker.C {
		c.refresh()
	}
}

func (c *cache) refresh() {
	startTs := time.Now()
	latestPosition, err := c.db.getLatestPosition()
	if err != nil {
		c.log.Error("error getting latest position", zap.Error(err))
		return
	}

	positions, err := c.db.fetchPositions(latestPosition.ID - 1000)
	if err != nil {
		c.log.Error("error getting last 1000 positions", zap.Error(err))
		return
	}

	c.mutex.Lock()
	c.data = positions
	c.mutex.Unlock()

	c.log.Info("cache updated with last 1000 positions", zap.Duration("duration", time.Since(startTs)))
}

func (c *cache) Close() {
	c.ticker.Stop()
}

// getPositions returns the last 1000 positions from the cache
func (c *cache) getPositions() []position {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.data
}
