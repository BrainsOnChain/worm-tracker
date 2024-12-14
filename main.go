package main

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/brainsonchain/worm-tracker/server"
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

	// -------------------------------------------------------------------------
	// Start the server
	log.Info("starting server")

	serverErr := make(chan error)

	// Start the server
	go func() {
		server := server.NewServer(log, "8080")
		serverErr <- server.Start()
	}()

	err = <-serverErr
	if err != nil {
		log.Sugar().Fatalf("error starting server: %v", err)
	}
}
