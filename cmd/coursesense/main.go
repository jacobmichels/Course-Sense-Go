package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jacobmichels/Course-Sense-Go/config"
	"github.com/jacobmichels/Course-Sense-Go/firestore"
	"github.com/jacobmichels/Course-Sense-Go/notifier"
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

	cfg, err := config.ReadConfig()
	if err != nil {
		log.Panicf("failed to get config: %s", err)
	}

	webadvisorService, err := webadvisor.NewWebAdvisorSectionService()
	if err != nil {
		log.Panicf("failed to create WebAdvisorSectionService: %s", err)
	}

	firestoreService, err := firestore.NewFirestoreWatcherService(ctx, cfg.Firestore.ProjectID, cfg.Firestore.SectionCollectionID, cfg.Firestore.WatcherCollectionID, cfg.Firestore.CredentialsFilePath)
	if err != nil {
		log.Panicf("failed to create FirestoreWatcherService: %s", err)
	}

	emailNotifier := notifier.NewEmail(cfg.Smtp.Host, cfg.Smtp.Username, cfg.Smtp.Password, cfg.Smtp.From, cfg.Smtp.Port)

	register := register.NewRegister(webadvisorService, firestoreService)
	trigger := trigger.NewTrigger(webadvisorService, firestoreService, emailNotifier)

	go func() {
		log.Println("starting trigger ticker")
		ticker := time.NewTicker(5 * time.Minute)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Println("triggering webadvisor poll")
				err := trigger.Trigger(ctx)
				if err != nil {
					log.Printf("failure occured during trigger: %v\n", err)
				}
			}
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := server.NewServer(fmt.Sprintf(":%s", port), cfg.Auth.Username, cfg.Auth.Password, register, trigger)
	if err = srv.Start(ctx); err != nil {
		log.Panicf("Server failure: %s", err)
	}
}
