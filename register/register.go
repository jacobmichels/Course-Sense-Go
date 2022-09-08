package register

import (
	"context"
	"fmt"

	coursesense "github.com/jacobmichels/Course-Sense-Go"
)

// Register implements RegistrationService
var _ coursesense.RegistrationService = Register{}

type Register struct {
	sectionService coursesense.SectionService
	watcherService coursesense.WatcherService
}

func NewRegister(s coursesense.SectionService, w coursesense.WatcherService) Register {
	return Register{s, w}
}

func (r Register) Register(ctx context.Context, section coursesense.Section, watcher coursesense.Watcher) error {
	// Registration steps
	// 1. Ensure the section exists
	// 2. Use the watcher service to persist the watcher to the section

	exists, err := r.sectionService.Exists(ctx, section)
	if err != nil {
		return fmt.Errorf("failed to check if section exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("section %s does not exist", section)
	}

	if err := r.watcherService.AddWatcher(ctx, section, watcher); err != nil {
		return fmt.Errorf("failed to persist %s to %s", watcher, section)
	}

	return nil
}
