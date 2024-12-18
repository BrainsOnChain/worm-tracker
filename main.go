package main

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/brainsonchain/worm-tracker/src"
)

func main() {

	// -------------------------------------------------------------------------
	// Initialize the logger

	logConfig := zap.NewDevelopmentConfig()
	logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	log, err := logConfig.Build()
	if err != nil {
		log.Sugar().Fatalf("error creating logger: %v", err)
	}
	zap.ReplaceGlobals(log)
	log.Info("logger initialized")

	if err := run(log); err != nil {
		log.Sugar().Fatalf("error running application: %v", err)
	}
}

func run(log *zap.Logger) error {

	// -------------------------------------------------------------------------
	// Initialize the database
	log.Info("initializing database")

	db, err := src.NewDBManager("database.db")
	if err != nil {
		return fmt.Errorf("error initializing database: %w", err)
	}
	defer db.Close()

	if err := db.CreatePositionsTable(); err != nil {
		return fmt.Errorf("error creating positions table: %w", err)
	}

	// -------------------------------------------------------------------------
	// Run the fetcher

	go src.Run(db)

	// -------------------------------------------------------------------------
	// Start the server
	log.Info("starting server")

	serverErr := make(chan error)

	// Start the server
	go func() {
		server := src.NewServer(log, "8080", db)
		serverErr <- server.Start()
	}()

	if err = <-serverErr; err != nil {
		log.Sugar().Fatalf("error starting server: %v", err)
	}

	return nil
}
