package repository

import (
	"context"
	"errors"
	"log"

	coursesense "github.com/jacobmichels/Course-Sense-Go"
	"github.com/jacobmichels/Course-Sense-Go/config"
)

func New(ctx context.Context, cfg config.Database) (coursesense.Repository, error) {
	if cfg.Type == "firestore" {
		log.Println("creating firestore repository")
		return newFirestoreRepository(ctx, cfg.Firestore)
	} else if cfg.Type == "sqlite" {
		log.Println("creating sqlite repository")
		return newSQLiteRepository(ctx, cfg.SQLite)
	} else {
		return nil, errors.New("invalid database type")
	}
}
