package notifier

import (
	"context"

	coursesense "github.com/jacobmichels/Course-Sense-Go"
)

var _ coursesense.Notifier = Noop{}

type Noop struct {
}

func NewNoop() Noop {
	return Noop{}
}

func (n Noop) Notify(ctx context.Context, section coursesense.Section, watchers ...coursesense.Watcher) error {
	return nil
}
