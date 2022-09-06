package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/jacobmichels/Course-Sense-Go/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		log.Println("receieved interrupt, shutting down")
		cancel()
	}()

	if err := server.Run(ctx); err != nil {
		log.Fatalf("error running server: %s", err)
	}
}
