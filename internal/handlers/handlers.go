package handlers

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Database interface {
	Close() error
}

func Ping() func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Println("Ping request receieved")

		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Error writing Ping response: %s", err)
		}
	}
}

func Trigger(db Database) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		log.Println("Trigger request receieved")

		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Error writing Trigger response: %s", err)
		}
	}
}

func Notify(db Database) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		log.Println("Notify request receieved")

		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Error writing Notify response: %s", err)
		}
	}
}
