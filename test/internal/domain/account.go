package domain

import "context"

type Account struct {
	ID string `json:"id"`
	
	CreatedAt interface{} `json:"created_at"`
	
	Name string `json:"name"`
	
	
	User []string `json:"user"`
	
}

type AccountRepository interface {
	List(ctx context.Context, limit, offset int) ([]*Account, error)
	Get(ctx context.Context, id string) (*Account, error)
	Create(ctx context.Context, model *Account) (string, error)
	Update(ctx context.Context, id string, model *Account) error
	Delete(ctx context.Context, id string) error
}
