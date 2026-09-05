package main

import (
	"fmt"
	"os"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting %s in %s mode on %s:%d\n", cfg.AppName, cfg.AppEnv, cfg.AppHost, cfg.AppPort)
}
