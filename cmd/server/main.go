package main

import (
	"log/slog"
	"os"

	"github.com/Masachusets/stock_wise/internal/config"
)

func main() {
	cfg := config.FromFlags()

	if err := runApp(cfg); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
