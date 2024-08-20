package main

import (
	"fmt"
	"log/slog"
	"os"
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
	logLevel := DefaultLogLevel
	if debug {
		logLevel = slog.LevelDebug
	}
	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
