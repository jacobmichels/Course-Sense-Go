package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jacobmichels/Course-Sense-Go/internal/database"
	"github.com/jacobmichels/Course-Sense-Go/internal/webadvisor"
	"github.com/julienschmidt/httprouter"
)

func Run(ctx context.Context) error {
	config, err := readAppConfig()
	if err != nil {
		return fmt.Errorf("failed to get viper config: %w", err)
	}
	db, err := database.NewFirestoreClient(ctx, config.Firestore.ProjectID, config.Firestore.CollectionID, config.Firestore.CredentialsFilePath)
	if err != nil {
		return fmt.Errorf("failed to create Firestore client: %w", err)
	}
	wa, err := webadvisor.NewWebAdvisor()
	if err != nil {
		return fmt.Errorf("failed to create webadvisor service: %w", err)
	}

	r := httprouter.New()

	// Ping endpoint for health check
	r.GET("/ping", pingHandler())
	// Triggers the empty space check
	// This endpoint is meant to be called by a cron job
	r.GET("/trigger", triggerHandler(db, wa))
	// Registers notifications
	r.PUT("/register", registerHandler(db, wa))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := http.Server{Addr: ":" + port, Handler: r}

	log.Printf("Listening on port %s", port)

	errChan := make(chan error)
	go func() { errChan <- srv.ListenAndServe() }()
	select {
	case err := <-errChan:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
	case <-ctx.Done():
		log.Println("gracefully shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}
		log.Println("server shutdown complete")
	}

	return nil
}
