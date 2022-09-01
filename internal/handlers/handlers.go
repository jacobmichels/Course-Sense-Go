package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/jacobmichels/Course-Sense-Go/internal/types"
	"github.com/julienschmidt/httprouter"
)

type Database interface {
	Close() error
	GetSections(context.Context) ([]*types.CourseSection, error)
	UpdateSection(context.Context, string, *types.User) error
	GetOrCreate(context.Context, *types.CourseSection) (string, error)
	GetWatchers(ctx context.Context, id string) ([]types.User, error)
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

		sections, err := db.GetSections(r.Context())
		if err != nil {
			log.Printf("Error getting sections: %s", err)
		}

		if err := json.NewEncoder(w).Encode(sections); err != nil {
			log.Printf("Error encoding sections: %s", err)
		}
	}
}

type NotifyRequest struct {
	// Validate that Section does not contain any watchers
	Section types.CourseSection
	User    types.User
}

func Notify(db Database) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		log.Println("Notify request receieved")

		// requests contains the Section and User to add to the watchers list
		// get the section from firestore, creating if it doesn't exist
		// add the user to the watchers list of the section
		//   if the user is already in the list, return OK (idempotent)
		//   if the user is not in the list, add the user to the list and return OK

		var request NotifyRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Printf("Error decoding Notify request: %s", err)
			http.Error(w, "Error decoding Notify request", http.StatusBadRequest)
			return
		}

		if err := ValidateNotifyRequest(&request); err != nil {
			log.Printf("Error validating Notify request: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id, err := db.GetOrCreate(r.Context(), &request.Section)
		if err != nil {
			log.Printf("Error getting or creating section: %s", err)
			http.Error(w, "Error getting or creating section", http.StatusInternalServerError)
			return
		}

		watchers, err := db.GetWatchers(r.Context(), id)
		if err != nil {
			log.Printf("Error getting watchers: %s", err)
			http.Error(w, "Error getting watchers", http.StatusInternalServerError)
			return
		}

		for _, watcher := range watchers {
			if watcher.Email == request.User.Email {
				log.Printf("User %s is already in the watchers list", request.User.Email)
				if _, err := w.Write([]byte("OK")); err != nil {
					log.Printf("Error writing Notify response: %s", err)
				}
				return
			} else if watcher.Phone == request.User.Phone {
				log.Printf("User %s is already in the watchers list", request.User.Phone)
				if _, err := w.Write([]byte("OK")); err != nil {
					log.Printf("Error writing Notify response: %s", err)
				}
				return
			}
		}

		request.Section.Watchers = append(request.Section.Watchers, request.User)
		err = db.UpdateSection(r.Context(), id, &request.User)
		if err != nil {
			log.Printf("Error updating section: %s", err)
			http.Error(w, "Error updating section", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func ValidateNotifyRequest(req *NotifyRequest) error {
	if req.Section.Watchers != nil && len(req.Section.Watchers) > 0 {
		return errors.New("Watchers field cannot be populated")
	}

	if req.Section.CourseCode <= 999 && req.Section.CourseCode >= 9999 {
		return errors.New("CourseCode must be 4 digits")
	}
	if req.Section.Department == "" {
		return errors.New("Department cannot be empty")
	}
	if req.Section.SectionCode == "" {
		return errors.New("SectionCode cannot be empty")
	}
	if req.Section.Term == "" {
		return errors.New("Term cannot be empty")
	}

	if req.User.Email == "" && req.User.Phone == "" {
		return errors.New("User must have either an email or phone")
	}

	return nil
}
