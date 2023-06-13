package notifier

import (
	"context"

	coursesense "github.com/jacobmichels/Course-Sense-Go"
	"github.com/rs/zerolog/log"
)

var _ coursesense.Notifier = Noop{}

type Noop struct {
}

func NewNoop() Noop {
	return Noop{}
}

func (n Noop) Notify(ctx context.Context, section coursesense.Section, watchers ...coursesense.Watcher) error {
	log.Info().Str("section", section.String()).Int("watcher_count", len(watchers)).Msg("noop notifier called")
	return nil
}
