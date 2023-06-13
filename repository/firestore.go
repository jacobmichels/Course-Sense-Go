package repository

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	coursesense "github.com/jacobmichels/Course-Sense-Go"
	"github.com/jacobmichels/Course-Sense-Go/config"
	"google.golang.org/api/option"
)

var _ coursesense.Repository = FirestoreRepository{}

type FirestoreRepository struct {
	firestore *firestore.Client
	cfg       config.Firestore
}

type FirestoreWatcher struct {
	Watcher   coursesense.Watcher `json:"watcher"`
	SectionID string              `json:"sectionID"`
}

func newFirestoreRepository(ctx context.Context, cfg config.Firestore) (FirestoreRepository, error) {
	// Create a new Firestore client using application default credentials.
	if cfg.CredentialsFile == "" {
		client, err := firestore.NewClient(ctx, cfg.ProjectID)
		if err != nil {
			return FirestoreRepository{}, err
		}

		return FirestoreRepository{client, cfg}, nil
	}

	// Create a new Firestore client using supplied credentials file.
	client, err := firestore.NewClient(ctx, cfg.ProjectID, option.WithCredentialsFile(cfg.CredentialsFile))
	if err != nil {
		return FirestoreRepository{}, err
	}

	return FirestoreRepository{client, cfg}, nil
}

func (f FirestoreRepository) AddWatcher(ctx context.Context, section coursesense.Section, watcher coursesense.Watcher) error {
	// Steps:
	// 1. Retrieve the section document, creating it if it doesn't exist
	// 2. Inspect the current watchers. If the new watcher is already a watcher, stop and return a nil error
	// 3. Append the new watcher to the watchers array
	// 4. Update the document in the collection

	documents, err := f.firestore.Collection(f.cfg.SectionCollectionID).Where("Code", "==", section.Code).Where("Term", "==", section.Term).Where("Course.Code", "==", section.Course.Code).Where("Course.Department", "==", section.Course.Department).Documents(ctx).GetAll()
	if err != nil {
		return fmt.Errorf("failed to get matching section documents: %w", err)
	}

	if len(documents) > 1 {
		return errors.New("more than one matching document found, expected 0 or 1")
	}

	var sectionID string
	if len(documents) == 0 {
		ref, _, err := f.firestore.Collection(f.cfg.SectionCollectionID).Add(ctx, section)
		if err != nil {
			return fmt.Errorf("failed to add %s to collection: %w", section, err)
		}
		sectionID = ref.ID
	} else {
		sectionID = documents[0].Ref.ID
	}

	documents, err = f.firestore.Collection(f.cfg.WatcherCollectionID).Where("SectionID", "==", sectionID).Documents(ctx).GetAll()
	if err != nil {
		return fmt.Errorf("failed to get matching watcher documents: %w", err)
	}

	for _, document := range documents {
		var firestoreWatcher FirestoreWatcher
		err := document.DataTo(&firestoreWatcher)
		if err != nil {
			return fmt.Errorf("failed to deserialize watcher: %w", err)
		}

		if firestoreWatcher.Watcher.Email == watcher.Email && firestoreWatcher.Watcher.Phone == watcher.Phone {
			// Watcher already watching this section, nothing to do
			return nil
		}
	}

	newWatcher := FirestoreWatcher{Watcher: watcher, SectionID: sectionID}
	_, _, err = f.firestore.Collection(f.cfg.WatcherCollectionID).Add(ctx, newWatcher)
	if err != nil {
		return fmt.Errorf("failed to write new watcher to collection: %w", err)
	}

	return nil
}

func (f FirestoreRepository) GetWatchedSections(ctx context.Context) ([]coursesense.Section, error) {
	documents, err := f.firestore.Collection(f.cfg.SectionCollectionID).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all documents in sections collection: %w", err)
	}

	var results []coursesense.Section
	for _, document := range documents {
		var result coursesense.Section
		err = document.DataTo(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize document: %w", err)
		}

		results = append(results, result)
	}

	return results, nil
}

func (f FirestoreRepository) GetWatchers(ctx context.Context, section coursesense.Section) ([]coursesense.Watcher, error) {
	documents, err := f.firestore.Collection(f.cfg.SectionCollectionID).Where("Code", "==", section.Code).Where("Term", "==", section.Term).Where("Course.Code", "==", section.Course.Code).Where("Course.Department", "==", section.Course.Department).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get matching section documents: %w", err)
	}

	// sanity check, we should never have more than one matching document
	if len(documents) > 1 {
		return nil, errors.New("more than one matching document found, expected 0 or 1")
	}

	if len(documents) == 0 {
		return nil, errors.New("section not found in firestore")
	}

	sectionID := documents[0].Ref.ID

	documents, err = f.firestore.Collection(f.cfg.WatcherCollectionID).Where("SectionID", "==", sectionID).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get matching watcher documents: %w", err)
	}

	var results []coursesense.Watcher
	for _, document := range documents {
		var result FirestoreWatcher
		err = document.DataTo(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize document: %w", err)
		}

		results = append(results, result.Watcher)
	}

	return results, nil
}

func (f FirestoreRepository) Cleanup(ctx context.Context, section coursesense.Section) error {
	documents, err := f.firestore.Collection(f.cfg.SectionCollectionID).Where("Code", "==", section.Code).Where("Term", "==", section.Term).Where("Course.Code", "==", section.Course.Code).Where("Course.Department", "==", section.Course.Department).Documents(ctx).GetAll()
	if err != nil {
		return fmt.Errorf("failed to get matching section documents: %w", err)
	}

	// sanity check, we should never have more than one matching document
	if len(documents) > 1 {
		return errors.New("more than one matching document found, expected 0 or 1")
	}

	if len(documents) == 0 {
		return errors.New("section not found in firestore")
	}

	sectionID := documents[0].Ref.ID
	_, err = documents[0].Ref.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete section: %w", err)
	}

	documents, err = f.firestore.Collection(f.cfg.WatcherCollectionID).Where("SectionID", "==", sectionID).Documents(ctx).GetAll()
	if err != nil {
		return fmt.Errorf("failed to get matching watcher documents: %w", err)
	}

	for _, document := range documents {
		_, err = document.Ref.Delete(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete watcher: %w", err)
		}
	}

	return nil
}
