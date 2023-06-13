package trigger

import (
	"context"
	"fmt"
	"log"

	coursesense "github.com/jacobmichels/Course-Sense-Go"
)

// Trigger implements TriggerService
var _ coursesense.TriggerService = Trigger{}

type Trigger struct {
	sectionService coursesense.SectionService
	watcherService coursesense.Repository
	notifiers      []coursesense.Notifier
}

func NewTrigger(s coursesense.SectionService, w coursesense.Repository, n ...coursesense.Notifier) Trigger {
	return Trigger{s, w, n}
}

// This function triggers a poll of webadvisor
func (t Trigger) Trigger(ctx context.Context) error {
	// Trigger steps
	// 1. Get all watched sections from the watcher service
	// 2. Loop over the sections, checking the available capacity on each
	// 3. If availability is found, use the notifiers to notify the watchers for that section
	// 4. Remove said watchers once successfully notified

	sections, err := t.watcherService.GetWatchedSections(ctx)
	if err != nil {
		return fmt.Errorf("failed to get watched sections: %w", err)
	}

	if len(sections) == 0 {
		log.Println("No watched sections")
		return nil
	}

	for _, section := range sections {
		available, err := t.sectionService.GetAvailableSeats(ctx, section)
		if err != nil {
			return fmt.Errorf("failed to get available seats for %s: %w", section, err)
		}

		log.Printf("%d available seats found for %s", available, section)

		if available == 0 {
			continue
		}

		watchers, err := t.watcherService.GetWatchers(ctx, section)
		if err != nil {
			return fmt.Errorf("failed to get watchers for %s: %w", section, err)
		}

		for _, notifier := range t.notifiers {
			err := notifier.Notify(ctx, section, watchers...)
			if err != nil {
				return fmt.Errorf("failed to notify watchers for %s: %w", section, err)
			}
		}

		if err := t.watcherService.Cleanup(ctx, section); err != nil {
			return fmt.Errorf("failed to remove watchers from %s: %w", section, err)
		}
	}

	return nil
}
