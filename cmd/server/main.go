package main

import (
	"fmt"
	"log"
	"os"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	app := server.New(*cfg)

	addr := fmt.Sprintf("%s:%d", cfg.AppHost, cfg.AppPort)
	log.Printf("Starting %s in %s mode on %s", cfg.AppName, cfg.AppEnv, addr)

	if err := app.Listen(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
