//go:build debug
// +build debug

package main

import (
	"log/slog"
)

var DefaultLogLevel = slog.LevelDebug

// func init() {
// 	// Set the minimum level to Debug
// 	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
// 		Level: slog.LevelDebug,
// 	})))
// }
