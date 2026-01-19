package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"test/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockUserRepository struct {
	Data map[string]*domain.User
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	var results []*domain.User
	for _, v := range m.Data {
		results = append(results, v)
	}
	
	// Simple slicing for mock pagination
	if offset >= len(results) {
		return []*domain.User{}, nil
	}
	end := offset + limit
	if end > len(results) {
		end = len(results)
	}
	return results[offset:end], nil
}

func (m *MockUserRepository) Get(ctx context.Context, id string) (*domain.User, error) {
	if val, ok := m.Data[id]; ok {
		return val, nil
	}
	return nil, nil
}

func (m *MockUserRepository) Create(ctx context.Context, model *domain.User) (string, error) {
	id := "test-id"
	model.ID = id
	if m.Data == nil {
		m.Data = make(map[string]*domain.User)
	}
	m.Data[id] = model
	return id, nil
}

func (m *MockUserRepository) Update(ctx context.Context, id string, model *domain.User) error {
	m.Data[id] = model
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	delete(m.Data, id)
	return nil
}

func TestUserHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &MockUserRepository{Data: make(map[string]*domain.User)}
	handler := NewUserHandler(repo)
	r := gin.Default()

	r.GET("/user", handler.List)
	r.POST("/user", handler.Create)

	t.Run("Create", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := domain.User{}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonBody))
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("List", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/user?page=1&limit=10", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
