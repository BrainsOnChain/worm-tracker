package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	// Prometheus Metrics
	log.Info("initializing prometheus metrics")

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":9091", nil)

	// -------------------------------------------------------------------------
	// Initialize the database
	log.Info("initializing database")

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./worm-tracker.sqlite" // Fallback for local
	}

	db, err := src.NewDBManager(dbPath)
	if err != nil {
		return fmt.Errorf("error initializing database: %w", err)
	}
	defer db.Close()

	cleanSlate := os.Getenv("CLEAN_SLATE") == "true"
	if err := db.Initialize(cleanSlate); err != nil {
		return fmt.Errorf("error creating positions table: %w", err)
	}

	// -------------------------------------------------------------------------
	// Error Channel
	log.Info("initializing error channels")

	wormErr := make(chan error)
	serverErr := make(chan error)

	// -------------------------------------------------------------------------
	// Start the fetcher
	log.Info("starting fetcher")

	fetcher, err := src.NewBlockFetcher(log)
	if err != nil {
		return fmt.Errorf("error initializing fetcher: %w", err)
	}

	go func() {
		if err := src.Run(fetcher, db); err != nil {
			wormErr <- err
		}
	}()

	// -------------------------------------------------------------------------
	// Start the server
	log.Info("starting server")

	server := src.NewServer(log, "8080", db)
	go func() {
		if err := server.Start(); err != nil {
			serverErr <- err
		}
	}()

	select {
	case err := <-wormErr:
		return fmt.Errorf("error running worm: %w", err)
	case err := <-serverErr:
		return fmt.Errorf("error running server: %w", err)
	}
}
