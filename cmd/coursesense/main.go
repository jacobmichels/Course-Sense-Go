package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/jacobmichels/Course-Sense-Go/firestore"
	"github.com/jacobmichels/Course-Sense-Go/register"
	"github.com/jacobmichels/Course-Sense-Go/server"
	"github.com/jacobmichels/Course-Sense-Go/trigger"
	"github.com/jacobmichels/Course-Sense-Go/webadvisor"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		log.Println("received interrupt, shutting down")
		cancel()
	}()

	webadvisorService, err := webadvisor.NewWebAdvisorSectionService()
	if err != nil {
		log.Panicf("failed to create WebAdvisorSectionService: %s", err)
	}

	firestoreService, err := firestore.NewFirestoreWatcherService(ctx, "playground-351923", "sections", "watchers", "")
	if err != nil {
		log.Panicf("failed to create FirestoreWatcherService: %s", err)
	}

	register := register.NewRegister(webadvisorService, firestoreService)
	trigger := trigger.NewTrigger(webadvisorService, firestoreService)

	srv := server.NewServer(":8080", register, trigger)
	if err = srv.Start(ctx); err != nil {
		log.Panicf("Server failure: %s", err)
	}
}
