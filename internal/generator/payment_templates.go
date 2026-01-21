package generator

const MercadoPagoTemplate = `package payments

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"{{.ProjectName}}/internal/domain"
	"{{.ProjectName}}/internal/config"
	"github.com/gin-gonic/gin"
)

type MercadoPagoService struct {
	AccessToken string
	Repo        domain.{{.Payments.TransactionsColl | title}}Repository
}

func NewMercadoPagoService(repo domain.{{.Payments.TransactionsColl | title}}Repository) *MercadoPagoService {
	return &MercadoPagoService{
		AccessToken: config.GetMPAccessToken(),
		Repo:        repo,
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
	transaction := &domain.{{.Payments.TransactionsColl | title}}{
		Provider:  "mercadopago",
		Payload:   "notification received", // Simplified for this template
		CreatedAt: time.Now(),
	}

	_, err := s.Repo.Create(c.Request.Context(), transaction)
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

const StripeTemplate = `package payments

import (
	"net/http"
	"time"

	"{{.ProjectName}}/internal/domain"
	"{{.ProjectName}}/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/webhook"
)

type StripeService struct {
	SecretKey      string
	WebhookSecret  string
	Repo           domain.{{.Payments.TransactionsColl | title}}Repository
}

func NewStripeService(repo domain.{{.Payments.TransactionsColl | title}}Repository) *StripeService {
	key := config.GetStripeSecretKey()
	stripe.Key = key
	return &StripeService{
		SecretKey:     key,
		WebhookSecret: config.GetStripeWebhookSecret(),
		Repo:          repo,
	}
}

type PaymentRequest struct {
	Amount   int64  ` + "`" + `json:"amount"` + "`" + ` // in cents
	Currency string ` + "`" + `json:"currency"` + "`" + `
}

func (s *StripeService) CreatePaymentIntentHandler(c *gin.Context) {
	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(req.Amount),
		Currency: stripe.String(req.Currency),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client_secret": pi.ClientSecret,
		"id":            pi.ID,
	})
}

func (s *StripeService) HandleWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	// Verify signature
	sigHeader := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, sigHeader, s.WebhookSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Save transaction/event to Firestore (simplified)
	transaction := &domain.{{.Payments.TransactionsColl | title}}{
		Provider:  "stripe",
		Payload:   string(event.Type),
		Status:    string(event.Type),
		CreatedAt: time.Now(),
	}

	// Just logging relevant events for the example
	if event.Type == "payment_intent.succeeded" {
		transaction.Status = "paid"
	}

	_, err = s.Repo.Create(c.Request.Context(), transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
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

func GetStripeSecretKey() string {
	token := os.Getenv("STRIPE_SECRET_KEY")
	if token == "" {
		return "your_stripe_secret_key"
	}
	return token
}

func GetStripeWebhookSecret() string {
	token := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if token == "" {
		return "your_stripe_webhook_secret"
	}
	return token
}

func GetJWTSecret() string {
	token := os.Getenv("JWT_SECRET")
	if token == "" {
		return "your_jwt_secret_key"
	}
	return token
}
`
