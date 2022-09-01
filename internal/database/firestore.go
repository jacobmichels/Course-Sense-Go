package database

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

type FirestoreDatabase struct {
	client *firestore.Client
}

func NewFirestoreClient(ctx context.Context, projectID string, credentialsPath string) (*FirestoreDatabase, error) {
	// Create a new Firestore client using application default credentials.
	if credentialsPath == "" {
		client, err := firestore.NewClient(ctx, projectID)
		if err != nil {
			return nil, err
		}

		return &FirestoreDatabase{client}, nil
	}

	// Create a new Firestore client using supplied credentials file.
	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, err
	}

	return &FirestoreDatabase{client}, nil
}

func (fd *FirestoreDatabase) Close() error {
	return fd.client.Close()
}
