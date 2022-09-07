package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/jacobmichels/Course-Sense-Go/internal/database"
	"github.com/jacobmichels/Course-Sense-Go/internal/types"
	"github.com/jacobmichels/Course-Sense-Go/internal/webadvisor"
	"github.com/julienschmidt/httprouter"
)

func pingHandler() func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Println("Ping request receieved")

		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Error writing Ping response: %s", err)
		}
	}
}

func triggerHandler(db *database.FirestoreDatabase, wa *webadvisor.WebAdvisor, notifiers ...Notifier) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		log.Println("Trigger request receieved")

		if err := trigger(r.Context(), db, wa, notifiers...); err != nil {
			log.Printf("Error occured during slot check: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}

func trigger(ctx context.Context, db *database.FirestoreDatabase, wa *webadvisor.WebAdvisor, notifiers ...Notifier) error {
	sections, ids, err := db.GetSections(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sections")
	}

	if len(sections) == 0 {
		log.Println("No sections registered")
		return nil
	}

	for i, section := range sections {
		capacity, err := wa.GetSectionCapacity(ctx, section)
		if err != nil {
			log.Printf("failed to get section capacity: %s", err)
		}

		if capacity > 0 {
			log.Printf("Section %s %d %s %s has %d capacity\n", section.Department, section.CourseCode, section.SectionCode, section.Term, capacity)

			for _, notifier := range notifiers {
				err = notifier.Notify(ctx, section)
				if err != nil {
					log.Printf("%s failed to notify: %s", notifier.Name(), err)
				}
			}

			sectionID := ids[i]
			err = db.DeleteSection(ctx, sectionID)
			if err != nil {
				log.Printf("failed to delete section: %s", err)
			}
		}
	}

	return nil
}

type RegisterRequest struct {
	// Validate that Section does not contain any watchers
	Section types.CourseSection
	User    types.User
}

func (req *RegisterRequest) Validate() error {
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

func registerHandler(db *database.FirestoreDatabase, wa *webadvisor.WebAdvisor) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		log.Println("Register request receieved")

		var request RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Printf("Error decoding Register request: %s", err)
			http.Error(w, "Error decoding Register request", http.StatusBadRequest)
			return
		}

		if err := request.Validate(); err != nil {
			log.Printf("Error validating Register request: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := register(r.Context(), db, wa, &request); err != nil {
			log.Printf("Error registering notification: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func register(ctx context.Context, db *database.FirestoreDatabase, wa *webadvisor.WebAdvisor, req *RegisterRequest) error {
	_, err := wa.GetSectionCapacity(ctx, &req.Section)
	if err != nil {
		return fmt.Errorf("failed to find requested section")
	}

	log.Printf("Section %s %d %s %s exists in WebAdvisor, adding watcher\n", req.Section.Department, req.Section.CourseCode, req.Section.SectionCode, req.Section.Term)

	id, err := db.GetOrCreate(ctx, &req.Section)
	if err != nil {
		return fmt.Errorf("Error getting or creating section: %w", err)
	}

	watchers, err := db.GetWatchers(ctx, id)
	if err != nil {
		return fmt.Errorf("Error getting watchers: %w", err)
	}

	for _, watcher := range watchers {
		if watcher.Email == req.User.Email {
			log.Printf("%s is already in the watchers list", req.User.Email)
			return nil
		} else if watcher.Phone == req.User.Phone {
			log.Printf("%s is already in the watchers list", req.User.Phone)
			return nil
		}
	}

	req.Section.Watchers = append(req.Section.Watchers, req.User)
	err = db.UpdateSection(ctx, id, &req.User)
	if err != nil {
		return fmt.Errorf("Error updating section: %w", err)
	}

	return nil
}
