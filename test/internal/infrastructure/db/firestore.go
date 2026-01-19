package db

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
	"google.golang.org/api/iterator"
)

// Repository defines the interface for database operations
type Repository interface {
	List(ctx context.Context, collection string) ([]map[string]interface{}, error)
	Get(ctx context.Context, collection, id string) (map[string]interface{}, error)
	Create(ctx context.Context, collection string, data map[string]interface{}) (string, error)
	Update(ctx context.Context, collection, id string, data map[string]interface{}) error
	Delete(ctx context.Context, collection, id string) error
	Close()
	GetClient() *firestore.Client
}

// FirestoreRepository implements Repository for Firestore
type FirestoreRepository struct {
	client *firestore.Client
}

// NewFirestoreRepository initializes the Firestore client and returns a Repository
func NewFirestoreRepository() (Repository, error) {
	ctx := context.Background()
	
	// Use credentials file copied to the project root
	opt := option.WithCredentialsFile("firebaseCredentials.json")
	conf := &firebase.Config{ProjectID: "tiendaonline-mvp"}
	
	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing firestore: %v", err)
	}

	return &FirestoreRepository{client: client}, nil
}

func (r *FirestoreRepository) Close() {
	if r.client != nil {
		r.client.Close()
	}
}

func (r *FirestoreRepository) GetClient() *firestore.Client {
	return r.client
}

func (r *FirestoreRepository) List(ctx context.Context, collection string) ([]map[string]interface{}, error) {
	iter := r.client.Collection(collection).Documents(ctx)
	var results []map[string]interface{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		data := doc.Data()
		data["id"] = doc.Ref.ID
		results = append(results, data)
	}
	return results, nil
}

func (r *FirestoreRepository) Get(ctx context.Context, collection, id string) (map[string]interface{}, error) {
	doc, err := r.client.Collection(collection).Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	data := doc.Data()
	data["id"] = doc.Ref.ID
	return data, nil
}

func (r *FirestoreRepository) Create(ctx context.Context, collection string, data map[string]interface{}) (string, error) {
	ref, _, err := r.client.Collection(collection).Add(ctx, data)
	if err != nil {
		return "", err
	}
	return ref.ID, nil
}

func (r *FirestoreRepository) Update(ctx context.Context, collection, id string, data map[string]interface{}) error {
	_, err := r.client.Collection(collection).Doc(id).Set(ctx, data, firestore.MergeAll)
	return err
}

func (r *FirestoreRepository) Delete(ctx context.Context, collection, id string) error {
	_, err := r.client.Collection(collection).Doc(id).Delete(ctx)
	return err
}
