package domain

import "context"

type User struct {
	ID string `json:"id"`
	
	CreatedAt interface{} `json:"created_at"`
	
	Email string `json:"email"`
	
	Name string `json:"name"`
	
	Picture string `json:"picture"`
	
	RoleId string `json:"roleId"`
	
	SettingsId string `json:"settingsId"`
	
	Uid string `json:"uid"`
	
	UpdatedAt interface{} `json:"updated_at"`
	
	
}

type UserRepository interface {
	List(ctx context.Context, limit, offset int) ([]*User, error)
	Get(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, model *User) (string, error)
	Update(ctx context.Context, id string, model *User) error
	Delete(ctx context.Context, id string) error
}
