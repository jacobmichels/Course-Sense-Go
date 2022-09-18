package coursesense

import (
	"context"
	"errors"
	"fmt"
)

// Domain types are defined in this file

type Course struct {
	Department string `json:"department"`
	Code       int    `json:"code"`
}

func (c Course) Valid() error {
	if c.Department == "" {
		return errors.New("Department cannot be empty")
	}
	if c.Code <= 999 && c.Code >= 9999 {
		return errors.New("Course code must be 4 digits")
	}

	return nil
}

type Section struct {
	Course Course `json:"course"`
	Code   string `json:"code"`
	Term   string `json:"term"`
}

func (s Section) Valid() error {
	if s.Code == "" {
		return errors.New("Section code cannot be empty")
	}
	if s.Term == "" {
		return errors.New("Term cannot be empty")
	}
	return s.Course.Valid()
}

func (s Section) String() string {
	return fmt.Sprintf("%s*%d*%s*%s", s.Course.Department, s.Course.Code, s.Code, s.Term)
}

// Service that gets information on course sections
type SectionService interface {
	Exists(context.Context, Section) (bool, error)
	GetAvailableSeats(context.Context, Section) (uint, error)
}

// A user registered for notifications on a Section
type Watcher struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

func (w Watcher) Valid() error {
	if w.Email == "" && w.Phone == "" {
		return errors.New("At least one contact method needs to be present")
	}

	return nil
}

func (w Watcher) String() string {
	return fmt.Sprintf("%s:%s", w.Email, w.Phone)
}

// Service that manages Watchers
type WatcherService interface {
	AddWatcher(context.Context, Section, Watcher) error
	GetWatchedSections(context.Context) ([]Section, error)
	GetWatchers(context.Context, Section) ([]Watcher, error)
	RemoveWatchers(context.Context, Section) error
}

// A type that sends can send notifications to Watchers
type Notifier interface {
	Notify(context.Context, Section, ...Watcher) error
}

type TriggerService interface {
	Trigger(context.Context) error
}

type RegistrationService interface {
	Register(context.Context, Section, Watcher) error
}
