package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func Ping(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("Ping request receieved")

	w.Write([]byte("OK"))
}

func Trigger(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("Trigger request receieved")

	w.Write([]byte("OK"))
}

func Notify(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("Notify request receieved")

	w.Write([]byte("OK"))
}
