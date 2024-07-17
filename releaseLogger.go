//go:build !debug
// +build !debug

package main

import "log/slog"

var DefaultLogLevel = slog.LevelInfo

// func init() {
// 	// Set the minimum level to Info (default behavior)
// 	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
// 		Level: slog.LevelInfo,
// 	})))
// }
