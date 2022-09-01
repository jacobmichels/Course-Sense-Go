package database

import (
	"context"

	"cloud.google.com/go/firestore"
)

type FirestoreDatabase struct {
	client *firestore.Client
}

func NewFirestoreClient(ctx context.Context, projectID string) (*FirestoreDatabase, error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &FirestoreDatabase{client}, nil
}

func (fd *FirestoreDatabase) Close() error {
	return fd.client.Close()
}
