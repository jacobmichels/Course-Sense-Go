package database

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/jacobmichels/Course-Sense-Go/internal/types"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type FirestoreDatabase struct {
	client       *firestore.Client
	collectionID string
}

func NewFirestoreClient(ctx context.Context, projectID, collectionID, credentialsPath string) (*FirestoreDatabase, error) {
	// Create a new Firestore client using application default credentials.
	if credentialsPath == "" {
		client, err := firestore.NewClient(ctx, projectID)
		if err != nil {
			return nil, err
		}

		return &FirestoreDatabase{client, collectionID}, nil
	}

	// Create a new Firestore client using supplied credentials file.
	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, err
	}

	return &FirestoreDatabase{client, collectionID}, nil
}

func (fd *FirestoreDatabase) Close() error {
	return fd.client.Close()
}

// Returns a list of all sections
func (fd *FirestoreDatabase) GetSections(ctx context.Context) ([]*types.CourseSection, error) {
	var sections []*types.CourseSection
	iter := fd.client.Collection(fd.collectionID).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var section *types.CourseSection
		err = doc.DataTo(&section)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal section: %v", err)
		}
		sections = append(sections, section)
	}

	return sections, nil
}

func (fd *FirestoreDatabase) UpdateSection(ctx context.Context, id string, user *types.User) error {
	_, err := fd.client.Collection(fd.collectionID).Doc(id).Update(ctx, []firestore.Update{
		{Path: "Watchers", Value: firestore.ArrayUnion(user)},
	})
	if err != nil {
		return err
	}

	return nil
}

// Gets a section, creating it if it doesn't exist. Returns the ID of the section
func (fd *FirestoreDatabase) GetOrCreate(ctx context.Context, section *types.CourseSection) (string, error) {
	documents, err := fd.client.Collection(fd.collectionID).Where("CourseCode", "==", section.CourseCode).Where("Department", "==", section.Department).Where("SectionCode", "==", section.SectionCode).Where("Term", "==", section.Term).Documents(ctx).GetAll()
	if err != nil {
		return "", fmt.Errorf("failed to get section: %w", err)
	}

	if len(documents) > 1 {
		return "", fmt.Errorf("found multiple sections with matching criteria")
	} else if len(documents) == 0 {
		// section does not exist, create it
		// sanity check, should not happen
		if section.Watchers != nil || len(section.Watchers) != 0 {
			return "", errors.New("refusing to create section with watchers")
		}

		ref, _, err := fd.client.Collection(fd.collectionID).Add(ctx, section)
		if err != nil {
			return "", fmt.Errorf("failed to create section: %w", err)
		}

		return ref.ID, nil
	}

	// section exists, return its ID
	return documents[0].Ref.ID, nil
}

func (fd *FirestoreDatabase) GetWatchers(ctx context.Context, id string) ([]types.User, error) {
	document, err := fd.client.Collection(fd.collectionID).Doc(id).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get section: %w", err)
	}

	var section types.CourseSection
	err = document.DataTo(&section)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal section: %v", err)
	}

	return section.Watchers, nil
}
