package main

import (
	"errors"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	client "github.com/7574-sistemas-distribuidos/tp-nivelador/src/client"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
)

func loadConfig() (client.ClientConfig, error) {
	agencyId := os.Getenv("AGENCY_ID")
	if agencyId == "" {
		return client.ClientConfig{}, errors.New("AGENCY_ID environment variable is required")
	}

	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		return client.ClientConfig{}, errors.New("SERVER_HOST environment variable is required")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		return client.ClientConfig{}, errors.New("SERVER_PORT environment variable is required")
	}

	inputFile := os.Getenv("INPUT_FILE")
	if inputFile == "" {
		return client.ClientConfig{}, errors.New("INPUT_FILE environment variable is required")
	}

	outputFile := os.Getenv("OUTPUT_FILE")
	if outputFile == "" {
		return client.ClientConfig{}, errors.New("OUTPUT_FILE environment variable is required")
	}

	batchSizeRaw := os.Getenv("BATCH_SIZE")
	if batchSizeRaw == "" {
		return client.ClientConfig{}, errors.New("BATCH_SIZE environment variable is required")
	}

	batchSize, err := strconv.Atoi(batchSizeRaw)
	if err != nil || batchSize <= 0 {
		return client.ClientConfig{}, errors.New("BATCH_SIZE must be a positive integer")
	}

	return client.ClientConfig{
		ServerHost: serverHost,
		ServerPort: serverPort,
		AgencyId:   agencyId,
		InputFile:  inputFile,
		OutputFile: outputFile,
		BatchSize:  batchSize,
	}, nil
}

func run() int {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)
	defer signal.Stop(signals)

	config, err := loadConfig()
	if err != nil {
		logger.Error("load-config", logger.Fail, "err", err)
		return 1
	}

	client, err := client.NewClient(config)
	if err != nil {
		select {
		case <-signals:
			return 0
		default:
		}
		logger.Error("client-new", logger.Fail, "err", err)
		return 1
	}

	done := make(chan error, 1)
	go func() { done <- client.Run() }()

	select {
	case <-signals:
		// Cerrar el socket desbloquea Read y Write
		// espero los defer de Run
		if err := client.Stop(); err != nil {
			logger.Warn("client-stop", logger.Fail, "err", err)
		}
		<-done
		return 0
	case err = <-done:
	}
	if err != nil {
		logger.Error("client-run", logger.Fail, "err", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
