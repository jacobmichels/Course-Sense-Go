package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func Ping(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("Ping request receieved")

	if _, err := w.Write([]byte("OK")); err != nil {
		log.Printf("Error writing Ping response: %s", err)
	}
}

func Trigger(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("Trigger request receieved")

	if _, err := w.Write([]byte("OK")); err != nil {
		log.Printf("Error writing Trigger response: %s", err)
	}
}

func Notify(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("Notify request receieved")

	if _, err := w.Write([]byte("OK")); err != nil {
		log.Printf("Error writing Notify response: %s", err)
	}
}
