package db

import (
	"context"
	"test/internal/domain"
	"google.golang.org/api/iterator"
)

type AccountRepository struct {
	client *FirestoreRepository
}

func NewAccountRepository(client *FirestoreRepository) *AccountRepository {
	return &AccountRepository{client: client}
}

func (r *AccountRepository) List(ctx context.Context, limit, offset int) ([]*domain.Account, error) {
	iter := r.client.client.Collection("account").Offset(offset).Limit(limit).Documents(ctx)
	var results []*domain.Account
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var m domain.Account
		if err := doc.DataTo(&m); err != nil {
			return nil, err
		}
		m.ID = doc.Ref.ID
		results = append(results, &m)
	}
	return results, nil
}

func (r *AccountRepository) Get(ctx context.Context, id string) (*domain.Account, error) {
	doc, err := r.client.client.Collection("account").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	var m domain.Account
	if err := doc.DataTo(&m); err != nil {
		return nil, err
	}
	m.ID = doc.Ref.ID
	return &m, nil
}

func (r *AccountRepository) Create(ctx context.Context, m *domain.Account) (string, error) {
	ref, _, err := r.client.client.Collection("account").Add(ctx, m)
	if err != nil {
		return "", err
	}
	return ref.ID, nil
}

func (r *AccountRepository) Update(ctx context.Context, id string, m *domain.Account) error {
	_, err := r.client.client.Collection("account").Doc(id).Set(ctx, m)
	return err
}

func (r *AccountRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.client.Collection("account").Doc(id).Delete(ctx)
	return err
}
