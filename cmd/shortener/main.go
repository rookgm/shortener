package main

import (
	"github.com/rookgm/shortener/config"
	"github.com/rookgm/shortener/internal/server"
	"log"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("Cannot initialize config: %v\n", err)
	}

	if err := server.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
