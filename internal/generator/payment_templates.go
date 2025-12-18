package generator

const MercadoPagoTemplate = `package payments

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"{{.ProjectName}}/internal/db"
	"{{.ProjectName}}/internal/config"
	"github.com/gin-gonic/gin"
)

type MercadoPagoService struct {
	AccessToken string
	Repo        db.Repository
	Collection  string
}

func NewMercadoPagoService(repo db.Repository, collection string) *MercadoPagoService {
	return &MercadoPagoService{
		AccessToken: config.GetMPAccessToken(),
		Repo:        repo,
		Collection:  collection,
	}
}

type PreferenceRequest struct {
	Items []Item ` + "`" + `json:"items"` + "`" + `
}

type Item struct {
	Title     string  ` + "`" + `json:"title"` + "`" + `
	Quantity  int     ` + "`" + `json:"quantity"` + "`" + `
	UnitPrice float64 ` + "`" + `json:"unit_price"` + "`" + `
}

func (s *MercadoPagoService) CreatePreference(ctx context.Context, req PreferenceRequest) (map[string]interface{}, error) {
	url := "https://api.mercadopago.com/checkout/preferences"
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.AccessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *MercadoPagoService) HandleWebhook(c *gin.Context) {
	var notification map[string]interface{}
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Save transaction to Firestore
	transaction := map[string]interface{}{
		"provider":   "mercadopago",
		"payload":    notification,
		"created_at": time.Now(),
	}

	_, err := s.Repo.Create(c.Request.Context(), s.Collection, transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

func (s *MercadoPagoService) CreatePreferenceHandler(c *gin.Context) {
	var req PreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := s.CreatePreference(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
`

const ConfigTemplate = `package config

import "os"

func GetMPAccessToken() string {
	token := os.Getenv("MP_ACCESS_TOKEN")
	if token == "" {
		return "YOUR_MERCADO_PAGO_ACCESS_TOKEN_HERE"
	}
	return token
}
`
