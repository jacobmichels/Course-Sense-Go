package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func PingHandler() func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Println("Ping request receieved")

		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Error writing Ping response: %s", err)
		}
	}
}

func TriggerHandler(db Database) func(http.ResponseWriter, *http.Request, httprouter.Params) {
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

type RegisterRequest struct {
	// Validate that Section does not contain any watchers
	Section types.CourseSection
	User    types.User
}

func RegisterHandler(db Database) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		log.Println("Register request receieved")

		var request RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Printf("Error decoding Register request: %s", err)
			http.Error(w, "Error decoding Register request", http.StatusBadRequest)
			return
		}

		if err := ValidateRegisterRequest(&request); err != nil {
			log.Printf("Error validating Register request: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := register(r.Context(), db, &request); err != nil {
			log.Printf("Error registering user: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func register(ctx context.Context, db Database, req *RegisterRequest) error {
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

func ValidateRegisterRequest(req *RegisterRequest) error {
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
