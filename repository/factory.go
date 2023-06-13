package repository

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"

	coursesense "github.com/jacobmichels/Course-Sense-Go"
	"github.com/jacobmichels/Course-Sense-Go/config"
)

func New(ctx context.Context, cfg config.Database) (coursesense.Repository, error) {
	if cfg.Type == "firestore" {
		log.Info().Msg("using firestore repository")
		return newFirestoreRepository(ctx, cfg.Firestore)
	} else if cfg.Type == "sqlite" {
		log.Info().Msg("using sqlite repository")
		return newSQLiteRepository(ctx, cfg.SQLite)
	} else {
		return nil, errors.New("invalid database type")
	}
}
