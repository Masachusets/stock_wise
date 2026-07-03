package main

import (
	"fmt"
	"os"

	"github.com/Masachusets/stock_wise/internal/config"
)

func main() {
	cfg := config.FromFlags()

	if err := runApp(cfg); err != nil {
		fmt.Printf("server failed: %v", err)
		os.Exit(1)
	}
}