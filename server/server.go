package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	coursesense "github.com/jacobmichels/Course-Sense-Go"
	"github.com/julienschmidt/httprouter"
)

type Server struct {
	registrationService coursesense.RegistrationService
	triggerService      coursesense.TriggerService
	addr                string
}

func NewServer(addr string, r coursesense.RegistrationService, t coursesense.TriggerService) Server {
	return Server{r, t, addr}
}

func (s Server) Start(ctx context.Context) error {
	r := httprouter.New()

	// register routes
	r.GET("/ping", s.pingHandler())
	r.PUT("/register", s.registerHandler())

	srv := http.Server{Addr: s.addr, Handler: r}
	log.Printf("listening on %s", s.addr)

	// start server, respecting context cancelation
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

func (s Server) pingHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Println("Ping request received")

		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Error writing ping response: %s", err)
		}
	}
}

type RegisterRequest struct {
	Section coursesense.Section `json:"section"`
	Watcher coursesense.Watcher `json:"watcher"`
}

func (r RegisterRequest) Valid() error {
	err := r.Section.Valid()
	if err != nil {
		return err
	}

	return r.Watcher.Valid()
}

func (s Server) registerHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Println("Register request received")

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("error decoding register request: %s", err)
			http.Error(w, "Failed to parse request", http.StatusBadRequest)
			return
		}

		if err := req.Valid(); err != nil {
			log.Printf("register request invalid: %s", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := s.registrationService.Register(r.Context(), req.Section, req.Watcher); err != nil {
			log.Printf("registration failed: %s", err)
			http.Error(w, "Registration failed, please ensure the course you are registering for exists. If error persists please contact service owner", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte("Registered for section\n")); err != nil {
			log.Printf("error writing register response: %s", err)
		}
		log.Printf("Register request succeeded: %s*%d*%s*%s for %s", req.Section.Course.Department, req.Section.Course.Code, req.Section.Code, req.Section.Term, req.Watcher.Email)
	}
}
