package main

import (
	"github.com/rookgm/shortener/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		panic(err)
	}
}
