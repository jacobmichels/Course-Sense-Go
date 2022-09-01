package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

func main() {
	r := httprouter.New()

	// Ping endpoint for health check
	r.GET("/ping", Ping)
	// Triggers the empty space check
	// This endpoint is meant to be called by a cron job
	r.GET("/trigger", Trigger)
	// Registers notifications
	r.PUT("/notify", Notify)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}
