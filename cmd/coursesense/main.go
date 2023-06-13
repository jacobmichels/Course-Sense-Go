package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jacobmichels/Course-Sense-Go/config"
	"github.com/jacobmichels/Course-Sense-Go/notifier"
	"github.com/jacobmichels/Course-Sense-Go/register"
	"github.com/jacobmichels/Course-Sense-Go/repository"
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

	cfg, err := config.ParseConfig()
	if err != nil {
		log.Panicf("failed to get config: %s", err)
	}

	webadvisorService, err := webadvisor.NewWebAdvisorSectionService()
	if err != nil {
		log.Panicf("failed to create WebAdvisorSectionService: %s", err)
	}

	repository, err := repository.New(ctx, cfg.Database)
	if err != nil {
		log.Panicf("failed to create repository: %s", err)
	}

	emailNotifier := notifier.NewEmail(cfg.Notifications.EmailSmtp.Host, cfg.Notifications.EmailSmtp.Username, cfg.Notifications.EmailSmtp.Password, cfg.Notifications.EmailSmtp.From, cfg.Notifications.EmailSmtp.Port)

	register := register.NewRegister(webadvisorService, repository)
	trigger := trigger.NewTrigger(webadvisorService, repository, emailNotifier)

	go func() {
		log.Printf("starting trigger ticker: polling every %d seconds\n", cfg.PollIntervalSecs)
		ticker := time.NewTicker(time.Second * time.Duration(cfg.PollIntervalSecs))

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

	srv := server.NewServer(fmt.Sprintf(":%s", port), register, trigger)
	if err = srv.Start(ctx); err != nil {
		log.Panicf("Server failure: %s", err)
	}
}
