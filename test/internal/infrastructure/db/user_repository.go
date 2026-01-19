package db

import (
	"context"
	"test/internal/domain"
	"google.golang.org/api/iterator"
)

type UserRepository struct {
	client *FirestoreRepository
}

func NewUserRepository(client *FirestoreRepository) *UserRepository {
	return &UserRepository{client: client}
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	iter := r.client.client.Collection("User").Offset(offset).Limit(limit).Documents(ctx)
	var results []*domain.User
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var m domain.User
		if err := doc.DataTo(&m); err != nil {
			return nil, err
		}
		m.ID = doc.Ref.ID
		results = append(results, &m)
	}
	return results, nil
}

func (r *UserRepository) Get(ctx context.Context, id string) (*domain.User, error) {
	doc, err := r.client.client.Collection("User").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	var m domain.User
	if err := doc.DataTo(&m); err != nil {
		return nil, err
	}
	m.ID = doc.Ref.ID
	return &m, nil
}

func (r *UserRepository) Create(ctx context.Context, m *domain.User) (string, error) {
	ref, _, err := r.client.client.Collection("User").Add(ctx, m)
	if err != nil {
		return "", err
	}
	return ref.ID, nil
}

func (r *UserRepository) Update(ctx context.Context, id string, m *domain.User) error {
	_, err := r.client.client.Collection("User").Doc(id).Set(ctx, m)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.client.Collection("User").Doc(id).Delete(ctx)
	return err
}
