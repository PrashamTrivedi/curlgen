package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	DefaultLogLevel = slog.LevelInfo
	logger          *slog.Logger
	taskFile        string
	changedFiles    []string
	prompt          string
	model           string
	openAIKey       string
	anthropicKey    string
	exampleCall     string
	apiGatewayJSON  string
	executeCurl     bool
	debug           bool
	apiKey          string
	apiURL          string
)

func main() {
	cobra.OnInitialize(initLogger)
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error executing root command: %v\n", err)
		os.Exit(1)
	}
}

func initLogger() {
	logLevel := DefaultLogLevel
	if debug {
		logLevel = slog.LevelDebug
	}
	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)
}
