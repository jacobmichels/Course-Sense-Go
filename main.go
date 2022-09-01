package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jacobmichels/Course-Sense-Go/internal/config"
	"github.com/jacobmichels/Course-Sense-Go/internal/database"
	"github.com/jacobmichels/Course-Sense-Go/internal/handlers"
	"github.com/julienschmidt/httprouter"
)

func main() {
	ctx := context.Background()
	r := httprouter.New()

	config, err := config.ReadAppConfig()
	if err != nil {
		log.Fatalf("Error reading config: %s", err)
	}
	log.Println("App config read")

	db, err := database.NewFirestoreClient(ctx, config.Firestore.ProjectID, config.Firestore.CollectionID, config.Firestore.CredentialsFilePath)
	if err != nil {
		log.Fatalf("Error creating Firestore client: %s", err)
	}
	log.Println("Database client created")

	// Ping endpoint for health check
	r.GET("/ping", handlers.PingHandler())
	// Triggers the empty space check
	// This endpoint is meant to be called by a cron job
	r.GET("/trigger", handlers.TriggerHandler(db))
	// Registers notifications
	r.PUT("/register", handlers.RegisterHandler(db))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on port %s", port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}
