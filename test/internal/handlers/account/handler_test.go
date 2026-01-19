package account

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

type MockAccountRepository struct {
	Data map[string]*domain.Account
}

func (m *MockAccountRepository) List(ctx context.Context, limit, offset int) ([]*domain.Account, error) {
	var results []*domain.Account
	for _, v := range m.Data {
		results = append(results, v)
	}
	
	// Simple slicing for mock pagination
	if offset >= len(results) {
		return []*domain.Account{}, nil
	}
	end := offset + limit
	if end > len(results) {
		end = len(results)
	}
	return results[offset:end], nil
}

func (m *MockAccountRepository) Get(ctx context.Context, id string) (*domain.Account, error) {
	if val, ok := m.Data[id]; ok {
		return val, nil
	}
	return nil, nil
}

func (m *MockAccountRepository) Create(ctx context.Context, model *domain.Account) (string, error) {
	id := "test-id"
	model.ID = id
	if m.Data == nil {
		m.Data = make(map[string]*domain.Account)
	}
	m.Data[id] = model
	return id, nil
}

func (m *MockAccountRepository) Update(ctx context.Context, id string, model *domain.Account) error {
	m.Data[id] = model
	return nil
}

func (m *MockAccountRepository) Delete(ctx context.Context, id string) error {
	delete(m.Data, id)
	return nil
}

func TestAccountHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &MockAccountRepository{Data: make(map[string]*domain.Account)}
	handler := NewAccountHandler(repo)
	r := gin.Default()

	r.GET("/account", handler.List)
	r.POST("/account", handler.Create)

	t.Run("Create", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := domain.Account{}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/account", bytes.NewBuffer(jsonBody))
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("List", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/account?page=1&limit=10", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
