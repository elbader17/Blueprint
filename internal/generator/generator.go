package generator

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/eduardo/blueprint/internal/domain"
)

// Generate creates the API project based on the config
func Generate(config *domain.Config, outputDir string, fs domain.FileSystemPort, template domain.TemplatePort) error {
	projectPath := filepath.Join(outputDir, config.ProjectName)
	fmt.Printf("Creating project at %s\n", projectPath)

	if err := fs.MkdirAll(projectPath); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Cleanup old directories/files if they exist

	if err := createDirectories(projectPath, config, fs); err != nil {
		return err
	}

	if err := generateGoMod(projectPath, config, fs); err != nil {
		return err
	}

	if err := generateDatabase(projectPath, config, fs, template); err != nil {
		return err
	}

	if err := generateAuth(projectPath, config, fs, template); err != nil {
		return err
	}

	if err := generatePayments(projectPath, config, fs, template); err != nil {
		return err
	}

	for _, model := range config.Models {
		if err := generateModelDomain(projectPath, config, model, fs, template); err != nil {
			return err
		}
		if err := generateModelRepository(projectPath, config, model, fs, template); err != nil {
			return err
		}
		if err := generateModelHandlers(projectPath, config, model, fs, template); err != nil {
			return err
		}
		if err := generateModelHandlerTests(projectPath, config, model, fs, template); err != nil {
			return err
		}
	}

	if err := copyFirebaseCredentials(projectPath, fs); err != nil {
		fmt.Printf("Warning: firebaseCredentials.json not found or could not be copied: %v\n", err)
	}

	if err := generateMain(projectPath, config, fs, template); err != nil {
		return err
	}

	if err := generateScripts(projectPath, config, fs); err != nil {
		return err
	}

	if err := generateArchitectureDocs(projectPath, fs); err != nil {
		return err
	}

	if err := generateDocsPlaceholder(projectPath, config, fs); err != nil {
		return err
	}

	if err := generateMakefile(projectPath, config, fs, template); err != nil {
		return err
	}

	if err := generateEnvFile(projectPath, config, fs); err != nil {
		return err
	}

	if err := generateDockerfile(projectPath, config, fs, template); err != nil {
		return err
	}

	if err := generateDockerCompose(projectPath, config, fs, template); err != nil {
		return err
	}

	return nil
}

func generateDockerCompose(projectPath string, config *domain.Config, fs domain.FileSystemPort, template domain.TemplatePort) error {
	const dockerComposeTemplate = `version: '3.8'

services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      {{if eq .Database.Type "postgresql"}}
      - DATABASE_URL=postgres://user:password@postgres:5432/{{.ProjectName}}
      {{else if eq .Database.Type "mongodb"}}
      - DATABASE_URL=mongodb://mongo:27017
      {{else if eq .Database.Type "firestore"}}
      - FIRESTORE_PROJECT_ID={{.Database.ProjectID}}
      - GOOGLE_APPLICATION_CREDENTIALS=/app/firebaseCredentials.json
      {{end}}
      {{if and .Auth .Auth.Enabled}}
      - MOCK_AUTH=false
      {{end}}
      {{if and .Payments .Payments.Enabled}}
      {{if eq .Payments.Provider "mercadopago"}}
      - MP_ACCESS_TOKEN=your_token_here
      {{else if eq .Payments.Provider "stripe"}}
      - STRIPE_SECRET_KEY=your_stripe_secret_key
      - STRIPE_WEBHOOK_SECRET=your_stripe_webhook_secret
      {{end}}
      {{end}}
    depends_on:
      {{if eq .Database.Type "postgresql"}}
      - postgres
      {{else if eq .Database.Type "mongodb"}}
      - mongo
      {{end}}

  {{if eq .Database.Type "postgresql"}}
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: {{.ProjectName}}
    ports:
      - "5432:5432"
  {{else if eq .Database.Type "mongodb"}}
  mongo:
    image: mongo:6.0
    ports:
      - "27017:27017"
  {{end}}
`
	content, err := template.Render("docker-compose", dockerComposeTemplate, config)
	if err != nil {
		return err
	}
	return fs.WriteFile(filepath.Join(projectPath, "docker-compose.yml"), content)
}

func generateDockerfile(projectPath string, config *domain.Config, fs domain.FileSystemPort, template domain.TemplatePort) error {
	const dockerfileTemplate = `# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go

# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
# Copy .env if it exists (usually better to pass env vars in docker-compose, but useful for standalone)
COPY --from=builder /app/.env .
{{if eq .Database.Type "firestore"}}
COPY --from=builder /app/firebaseCredentials.json .
{{end}}

EXPOSE 8080
CMD ["./main"]
`
	content, err := template.Render("dockerfile", dockerfileTemplate, config)
	if err != nil {
		return err
	}
	return fs.WriteFile(filepath.Join(projectPath, "Dockerfile"), content)
}

func generateMakefile(projectPath string, config *domain.Config, fs domain.FileSystemPort, template domain.TemplatePort) error {
	const makefileTemplate = `PROJECT_ID ?= {{.Database.ProjectID}}
REGION ?= us-central1
SERVICE_NAME ?= {{.ProjectName}}
IMAGE_NAME ?= gcr.io/$(PROJECT_ID)/$(SERVICE_NAME)

.PHONY: run build test docker-build docker-push deploy

run:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

test:
	go test ./...

docker-build:
	docker build -t $(IMAGE_NAME) .

docker-push:
	docker push $(IMAGE_NAME)

deploy:
	gcloud run deploy $(SERVICE_NAME) \
		--image $(IMAGE_NAME) \
		--region $(REGION) \
		--platform managed \
		--allow-unauthenticated \
		--project $(PROJECT_ID)
`
	// Fallback if ProjectID is empty (e.g. for Postgres/Mongo if not specified)
	if config.Database.ProjectID == "" {
		config.Database.ProjectID = "your-project-id"
	}

	content, err := template.Render("makefile", makefileTemplate, config)
	if err != nil {
		return err
	}
	return fs.WriteFile(filepath.Join(projectPath, "Makefile"), content)
}

func createDirectories(projectPath string, config *domain.Config, fs domain.FileSystemPort) error {
	dirs := []string{
		"cmd/api",
		"internal/domain",
		"internal/infrastructure/db",
		"internal/auth",
		"internal/handlers/auth",
		"internal/payments",
		"internal/config",
	}
	for _, model := range config.Models {
		dirs = append(dirs, filepath.Join("internal/handlers", strings.ToLower(model.Name)))
	}
	for _, dir := range dirs {
		if err := fs.MkdirAll(filepath.Join(projectPath, dir)); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

func generateAuth(projectPath string, config *domain.Config, fs domain.FileSystemPort, template domain.TemplatePort) error {
	if config.Auth == nil || !config.Auth.Enabled {
		return nil
	}
	return generateAuthFiles(projectPath, config, fs, template)
}

func generatePayments(projectPath string, config *domain.Config, fs domain.FileSystemPort, template domain.TemplatePort) error {
	// Generate config if Auth or Payments enabled
	if (config.Auth != nil && config.Auth.Enabled) || (config.Payments != nil && config.Payments.Enabled) {
		if err := generateConfigFiles(projectPath, config, fs, template); err != nil {
			return err
		}
	}

	if config.Payments == nil || !config.Payments.Enabled {
		return nil
	}
	return generatePaymentFiles(projectPath, config, fs, template)
}

func copyFirebaseCredentials(projectPath string, fs domain.FileSystemPort) error {
	return fs.CopyFile("firebaseCredentials.json", filepath.Join(projectPath, "firebaseCredentials.json"))
}

func generateScripts(projectPath string, config *domain.Config, fs domain.FileSystemPort) error {
	if err := generateTestScript(projectPath, config, fs); err != nil {
		return err
	}
	if err := generateSetupScript(projectPath, config, fs); err != nil {
		return err
	}
	return generateDocsScript(projectPath, fs)
}

func generateGoMod(projectPath string, config *domain.Config, fs domain.FileSystemPort) error {
	var deps []string
	deps = append(deps, "github.com/gin-gonic/gin v1.9.1")
	deps = append(deps, "github.com/swaggo/files v1.0.1")
	deps = append(deps, "github.com/swaggo/gin-swagger v1.6.0")
	deps = append(deps, "github.com/swaggo/swag v1.16.2")
	deps = append(deps, "github.com/stretchr/testify v1.8.4")
	deps = append(deps, "github.com/joho/godotenv v1.5.1")

	if config.Auth != nil && config.Auth.Enabled {
		deps = append(deps, "firebase.google.com/go/v4 v4.13.0")
		if config.Auth.Provider == "jwt" {
			deps = append(deps, "github.com/golang-jwt/jwt/v5 v5.2.0")
			deps = append(deps, "golang.org/x/crypto v0.19.0")
		}
	}

	if config.Payments != nil && config.Payments.Enabled && config.Payments.Provider == "stripe" {
		deps = append(deps, "github.com/stripe/stripe-go/v76 v76.0.0")
	}

	switch config.Database.Type {
	case "firestore":
		deps = append(deps, "cloud.google.com/go/firestore v1.14.0")
		deps = append(deps, "google.golang.org/api v0.150.0")
	case "postgresql":
		deps = append(deps, "github.com/jackc/pgx/v5 v5.5.0")
	case "mongodb":
		deps = append(deps, "go.mongodb.org/mongo-driver v1.13.0")
	}

	content := fmt.Sprintf(`module %s

go 1.23

require (
	%s
)
`, config.ProjectName, strings.Join(deps, "\n\t"))
	return fs.WriteFile(filepath.Join(projectPath, "go.mod"), []byte(content))
}

func generateDatabase(projectPath string, config *domain.Config, fs domain.FileSystemPort, template domain.TemplatePort) error {
	switch config.Database.Type {
	case "firestore":
		return generateFirestore(projectPath, config, fs, template)
	case "postgresql":
		return generatePostgres(projectPath, config, fs, template)
	case "mongodb":
		return generateMongo(projectPath, config, fs, template)
	default:
		return fmt.Errorf("unsupported database type: %s", config.Database.Type)
	}
}

func generatePostgres(projectPath string, config *domain.Config, fs domain.FileSystemPort, template domain.TemplatePort) error {
	content, err := template.Render("postgres_base", PostgresBaseTemplate, config)
	if err != nil {
		return err
	}
	return fs.WriteFile(filepath.Join(projectPath, "internal/infrastructure/db/postgres.go"), content)
}

func generateMongo(projectPath string, config *domain.Config, fs domain.FileSystemPort, template domain.TemplatePort) error {
	content, err := template.Render("mongo_base", MongoBaseTemplate, config)
	if err != nil {
		return err
	}
	return fs.WriteFile(filepath.Join(projectPath, "internal/infrastructure/db/mongo.go"), content)
}

func generateFirestore(projectPath string, config *domain.Config, fs domain.FileSystemPort, template domain.TemplatePort) error {
	const firestoreTemplate = `package db

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
	conf := &firebase.Config{ProjectID: "{{.Database.ProjectID}}"}
	
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
`
	if config.FirestoreProjectID == "" {
		config.FirestoreProjectID = "tiendaonline-mvp"
	}

	content, err := template.Render("firestore", firestoreTemplate, config)
	if err != nil {
		return err
	}
	return fs.WriteFile(filepath.Join(projectPath, "internal/infrastructure/db/firestore.go"), content)
}

func generateModelDomain(projectPath string, config *domain.Config, model domain.Model, fs domain.FileSystemPort, template domain.TemplatePort) error {
	const domainTemplate = `package domain

import (
	"context"
	{{$hasTime := false}}{{range $k, $v := .Model.Fields}}{{if eq $v "datetime"}}{{$hasTime = true}}{{end}}{{end}}{{if $hasTime}}"time"{{end}}
)

type {{.Model.Name | title}} struct {
	ID string ` + "`" + `json:"id" bson:"_id,omitempty"` + "`" + `
	{{range $k, $v := .Model.Fields}}
	{{$k | pascal}} {{if eq $v "string"}}string{{else if eq $v "integer"}}int{{else if eq $v "float"}}float64{{else if eq $v "boolean"}}bool{{else if eq $v "datetime"}}time.Time{{else}}interface{}{{end}} ` + "`" + `json:"{{$k}}" bson:"{{$k}}"` + "`" + `
	{{end}}
	{{range $k, $v := .Model.Relations}}
	{{$k | pascal}} {{if hasPrefix $v "hasMany"}}[]string{{else}}string{{end}} ` + "`" + `json:"{{$k}}" bson:"{{$k}}"` + "`" + `
	{{end}}
}

type {{.Model.Name | title}}Repository interface {
	List(ctx context.Context, limit, offset int) ([]*{{.Model.Name | title}}, error)
	Get(ctx context.Context, id string) (*{{.Model.Name | title}}, error)
	Create(ctx context.Context, model *{{.Model.Name | title}}) (string, error)
	Update(ctx context.Context, id string, model *{{.Model.Name | title}}) error
	Delete(ctx context.Context, id string) error
}
`
	data := struct {
		ProjectName string
		Model       domain.Model
	}{
		ProjectName: config.ProjectName,
		Model:       model,
	}

	content, err := template.Render(model.Name+"_domain", domainTemplate, data)
	if err != nil {
		return err
	}
	return fs.WriteFile(filepath.Join(projectPath, "internal/domain", strings.ToLower(model.Name)+".go"), content)
}

func generateModelHandlers(projectPath string, config *domain.Config, model domain.Model, fs domain.FileSystemPort, template domain.TemplatePort) error {
	const handlerTemplate = `package {{.Model.Name | lower}}

import (
	"net/http"
	"strconv"
	"{{.ProjectName}}/internal/domain"
	"github.com/gin-gonic/gin"
)

type {{.Model.Name | title}}Handler struct {
	repo domain.{{.Model.Name | title}}Repository
}

func New{{.Model.Name | title}}Handler(repo domain.{{.Model.Name | title}}Repository) *{{.Model.Name | title}}Handler {
	return &{{.Model.Name | title}}Handler{repo: repo}
}

func (h *{{.Model.Name | title}}Handler) List(c *gin.Context) {
	limit := {{if .DefaultLimit}}{{.DefaultLimit}}{{else}}10{{end}}
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	page := 1
	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	offset := (page - 1) * limit

	results, err := h.repo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

func (h *{{.Model.Name | title}}Handler) Get(c *gin.Context) {
	id := c.Param("id")
	result, err := h.repo.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *{{.Model.Name | title}}Handler) Create(c *gin.Context) {
	var m domain.{{.Model.Name | title}}
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := h.repo.Create(c.Request.Context(), &m)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	m.ID = id
	c.JSON(http.StatusCreated, m)
}

func (h *{{.Model.Name | title}}Handler) Update(c *gin.Context) {
	id := c.Param("id")
	var m domain.{{.Model.Name | title}}
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.Update(c.Request.Context(), id, &m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *{{.Model.Name | title}}Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
`
	data := struct {
		ProjectName  string
		Model        domain.Model
		DefaultLimit int
	}{
		ProjectName:  config.ProjectName,
		Model:        model,
		DefaultLimit: 10,
	}
	if config.Pagination != nil && config.Pagination.DefaultLimit > 0 {
		data.DefaultLimit = config.Pagination.DefaultLimit
	}

	content, err := template.Render(model.Name+"_handler", handlerTemplate, data)
	if err != nil {
		return err
	}
	return fs.WriteFile(filepath.Join(projectPath, "internal/handlers", strings.ToLower(model.Name), "handler.go"), content)
}

func generatePaymentFiles(projectPath string, config *domain.Config, fs domain.FileSystemPort, template domain.TemplatePort) error {
	if config.Payments.Provider == "mercadopago" {
		content, err := template.Render("mercadopago", MercadoPagoTemplate, config)
		if err != nil {
			return err
		}
		return fs.WriteFile(filepath.Join(projectPath, "internal/payments/mercadopago.go"), content)
	} else if config.Payments.Provider == "stripe" {
		content, err := template.Render("stripe", StripeTemplate, config)
		if err != nil {
			return err
		}
		return fs.WriteFile(filepath.Join(projectPath, "internal/payments/stripe.go"), content)
	}
	return nil
}

func generateConfigFiles(projectPath string, config *domain.Config, fs domain.FileSystemPort, template domain.TemplatePort) error {
	content, err := template.Render("config", ConfigTemplate, config)
	if err != nil {
		return err
	}
	return fs.WriteFile(filepath.Join(projectPath, "internal/config/config.go"), content)
}

func generateAuthFiles(projectPath string, config *domain.Config, fs domain.FileSystemPort, template domain.TemplatePort) error {
	var middlewareTemplate string
	var handlerTemplateStr string

	if config.Auth.Provider == "jwt" {
		middlewareTemplate = JWTMiddlewareTemplate
		handlerTemplateStr = JWTAuthHandlerTemplate

		// Generate domain/auth.go for UserAuthData
		const authDomainTemplate = `package domain

import "context"

// UserAuthData represents minimal data needed for auth
type UserAuthData struct {
	ID       string
	Email    string
	Password string
	Role     string
}

// UserRepository interface to fetch password hashes
type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*UserAuthData, error)
	RegisterUser(ctx context.Context, user *UserAuthData) (string, error)
}
`
		if err := fs.WriteFile(filepath.Join(projectPath, "internal/domain/auth.go"), []byte(authDomainTemplate)); err != nil {
			return err
		}

	} else {
		middlewareTemplate = AuthMiddlewareTemplate
		handlerTemplateStr = AuthHandlerTemplate
	}

	middlewareContent, err := template.Render("auth_middleware", middlewareTemplate, config)
	if err != nil {
		return err
	}
	if err := fs.WriteFile(filepath.Join(projectPath, "internal/auth/middleware.go"), middlewareContent); err != nil {
		return err
	}

	content, err := template.Render("auth_handler", handlerTemplateStr, config)
	if err != nil {
		return err
	}
	return fs.WriteFile(filepath.Join(projectPath, "internal/handlers/auth/handler.go"), content)
}

func generateModelRepository(projectPath string, config *domain.Config, model domain.Model, fs domain.FileSystemPort, template domain.TemplatePort) error {
	var repoTemplate string
	switch config.Database.Type {
	case "firestore":
		repoTemplate = `package db

import (
	"context"
	{{if .IsJWT}}"time"{{end}}
	"{{.ProjectName}}/internal/domain"
	"google.golang.org/api/iterator"
)

type {{.Model.Name | title}}Repository struct {
	client *FirestoreRepository
}

func New{{.Model.Name | title}}Repository(client *FirestoreRepository) *{{.Model.Name | title}}Repository {
	return &{{.Model.Name | title}}Repository{client: client}
}

// GetByEmail is used for JWT auth
{{if .IsJWT}}
func (r *{{.Model.Name | title}}Repository) GetByEmail(ctx context.Context, email string) (*domain.UserAuthData, error) {
	iter := r.client.client.Collection("{{.Model.Name}}").Where("email", "==", email).Documents(ctx)
	doc, err := iter.Next()
	if err != nil {
		return nil, err
	}
	var m domain.{{.Model.Name | title}}
	if err := doc.DataTo(&m); err != nil {
		return nil, err
	}
	// Map to UserAuthData
	return &domain.UserAuthData{
		ID: doc.Ref.ID,
		Email: m.Email,
		{{if .IsJWT}}Password: m.Password,{{end}}
		Role: m.RoleId,
	}, nil
}
{{end}}

func (r *{{.Model.Name | title}}Repository) List(ctx context.Context, limit, offset int) ([]*domain.{{.Model.Name | title}}, error) {
	iter := r.client.client.Collection("{{.Model.Name}}").Offset(offset).Limit(limit).Documents(ctx)
	var results []*domain.{{.Model.Name | title}}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var m domain.{{.Model.Name | title}}
		if err := doc.DataTo(&m); err != nil {
			return nil, err
		}
		m.ID = doc.Ref.ID
		results = append(results, &m)
	}
	return results, nil
}

func (r *{{.Model.Name | title}}Repository) Get(ctx context.Context, id string) (*domain.{{.Model.Name | title}}, error) {
	doc, err := r.client.client.Collection("{{.Model.Name}}").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	var m domain.{{.Model.Name | title}}
	if err := doc.DataTo(&m); err != nil {
		return nil, err
	}
	m.ID = doc.Ref.ID
	return &m, nil
}

func (r *{{.Model.Name | title}}Repository) Create(ctx context.Context, model *domain.{{.Model.Name | title}}) (string, error) {
	ref, _, err := r.client.client.Collection("{{.Model.Name}}").Add(ctx, model)
	if err != nil {
		return "", err
	}
	return ref.ID, nil
}

{{if .IsJWT}}
func (r *{{.Model.Name | title}}Repository) RegisterUser(ctx context.Context, user *domain.UserAuthData) (string, error) {
	now := time.Now()
	data := map[string]interface{}{
		"email":      user.Email,
		"password":   user.Password,
		"role_id":    user.Role,
		"created_at": now,
		"updated_at": now,
	}
	ref, _, err := r.client.client.Collection("{{.Model.Name}}").Add(ctx, data)
	if err != nil {
		return "", err
	}
	return ref.ID, nil
}
{{end}}

func (r *{{.Model.Name | title}}Repository) Update(ctx context.Context, id string, m *domain.{{.Model.Name | title}}) error {
	_, err := r.client.client.Collection("{{.Model.Name}}").Doc(id).Set(ctx, m)
	return err
}

func (r *{{.Model.Name | title}}Repository) Delete(ctx context.Context, id string) error {
	_, err := r.client.client.Collection("{{.Model.Name}}").Doc(id).Delete(ctx)
	return err
}
`
	case "postgresql":
		// I'll need to adapt the Postgres template dynamically too, but let's stick to the base pattern first.
		// For now, I'll update the PostgresRepoTemplate constant if possible, or inject the GetByEmail logic.
		// Since PostgresRepoTemplate is likely a large string, I should check if I can modify it or append methods.
		// The current architecture uses a fixed template. I might need to make it more flexible.
		// For simplicity in this iteration, I will inject the GetByEmail method into the template data if it's the User collection.
		repoTemplate = PostgresRepoTemplate
	case "mongodb":
		repoTemplate = MongoRepoTemplate
	}
	var allFields []string
	for k := range model.Fields {
		allFields = append(allFields, k)
	}
	for k := range model.Relations {
		allFields = append(allFields, k)
	}
	sort.Strings(allFields)

	// Pre-calculate SQL parts for Postgres
	var insertCols []string
	var insertPlaceholders []string
	var updateSet []string
	var selectCols []string
	var schemaCols []string

	selectCols = append(selectCols, "id")
	// ID column definition
	schemaCols = append(schemaCols, "id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text")

	for i, f := range allFields {
		insertCols = append(insertCols, f)
		insertPlaceholders = append(insertPlaceholders, fmt.Sprintf("$%d", i+1))
		updateSet = append(updateSet, fmt.Sprintf("%s = $%d", f, i+1))
		selectCols = append(selectCols, f)

		// Map Go types to SQL types
		sqlType := "TEXT"
		if fieldType, ok := model.Fields[f]; ok {
			switch fieldType {
			case "int":
				sqlType = "INTEGER"
			case "float":
				sqlType = "DOUBLE PRECISION"
			case "bool":
				sqlType = "BOOLEAN"
			case "datetime":
				sqlType = "TIMESTAMP"
			}
		}
		// Check for relations
		if relationType, ok := model.Relations[f]; ok {
			if strings.HasPrefix(relationType, "hasMany") {
				sqlType = "TEXT[]"
			}
		}
		schemaCols = append(schemaCols, fmt.Sprintf("%s %s", f, sqlType))
	}

	isJWT := false
	if config.Auth != nil && config.Auth.Enabled && config.Auth.Provider == "jwt" {
		if strings.EqualFold(model.Name, config.Auth.UserCollection) {
			isJWT = true
		}
	}

	data := struct {
		ProjectName        string
		Model              domain.Model
		Fields             []string // For struct generation/scanning if needed
		InsertColumns      string
		InsertPlaceholders string
		UpdateSet          string
		SelectColumns      string
		CreateTableSQL     string
		TotalFields        int
		IsJWT              bool
	}{
		ProjectName:        config.ProjectName,
		Model:              model,
		Fields:             allFields,
		InsertColumns:      strings.Join(insertCols, ", "),
		InsertPlaceholders: strings.Join(insertPlaceholders, ", "),
		UpdateSet:          strings.Join(updateSet, ", "),
		SelectColumns:      strings.Join(selectCols, ", "),
		CreateTableSQL:     fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", model.Name, strings.Join(schemaCols, ", ")),
		TotalFields:        len(allFields),
		IsJWT:              isJWT,
	}

	content, err := template.Render(model.Name+"_repo", repoTemplate, data)
	if err != nil {
		return err
	}
	return fs.WriteFile(filepath.Join(projectPath, "internal/infrastructure/db", strings.ToLower(model.Name)+"_repository.go"), content)
}

func generateEnvFile(projectPath string, config *domain.Config, fs domain.FileSystemPort) error {
	var buffer bytes.Buffer

	buffer.WriteString("PORT=8080\n")

	if config.Database.Type == "postgresql" || config.Database.Type == "mongodb" {
		url := config.Database.URL
		if url == "" {
			url = "your_database_url_here"
		}
		buffer.WriteString(fmt.Sprintf("DATABASE_URL=%s\n", url))
	}

	if config.Auth != nil && config.Auth.Enabled && config.Auth.Provider == "jwt" {
		buffer.WriteString("JWT_SECRET=your_secret_key_here\n")
	}

	if config.Payments != nil && config.Payments.Enabled {
		if config.Payments.Provider == "mercadopago" {
			buffer.WriteString("MP_ACCESS_TOKEN=your_mercadopago_access_token_here\n")
		} else if config.Payments.Provider == "stripe" {
			buffer.WriteString("STRIPE_SECRET_KEY=your_stripe_secret_key_here\n")
			buffer.WriteString("STRIPE_WEBHOOK_SECRET=your_stripe_webhook_secret_here\n")
		}
	}

	return fs.WriteFile(filepath.Join(projectPath, ".env"), buffer.Bytes())
}

func generateMain(projectPath string, config *domain.Config, fs domain.FileSystemPort, template domain.TemplatePort) error {
	const mainTemplate = `package main

import (
	{{if and .Auth .Auth.Enabled}}{{if eq .Auth.Provider "firebase"}}"context"{{end}}{{end}}
	"log"
	"os"
	{{if not (and .Auth .Auth.Enabled)}}
	"net/http"
	"strings"
	{{end}}

	"github.com/joho/godotenv"
	"{{.ProjectName}}/internal/infrastructure/db"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "{{.ProjectName}}/docs"
	{{if and .Auth .Auth.Enabled}}
	authService "{{.ProjectName}}/internal/auth"
	authHandler "{{.ProjectName}}/internal/handlers/auth"
	{{if eq .Auth.Provider "firebase"}}firebase "firebase.google.com/go/v4"{{end}}
	{{end}}
	{{if and .Payments .Payments.Enabled}}
	"{{.ProjectName}}/internal/payments"
	{{end}}
	{{range .Models}}
	"{{$.ProjectName}}/internal/handlers/{{.Name | lower}}"
	{{end}}
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize Database
	{{if eq .Database.Type "firestore"}}
	baseRepo, err := db.NewFirestoreRepository()
	{{else if eq .Database.Type "postgresql"}}
	baseRepo, err := db.NewPostgresRepository(os.Getenv("DATABASE_URL"))
	{{else if eq .Database.Type "mongodb"}}
	baseRepo, err := db.NewMongoRepository(os.Getenv("DATABASE_URL"), "{{.ProjectName}}")
	{{end}}
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer baseRepo.Close()

	{{if and .Auth .Auth.Enabled}}
	// Initialize Auth Service
	{{if eq .Auth.Provider "jwt"}}
	// Initialize User Repo for JWT
	{{if eq .Database.Type "firestore"}}
	userRepo := db.New{{.Auth.UserCollection | title}}Repository(baseRepo.(*db.FirestoreRepository))
	{{else if eq .Database.Type "postgresql"}}
	userRepo := db.New{{.Auth.UserCollection | title}}Repository(baseRepo.(*db.PostgresRepository))
	{{else if eq .Database.Type "mongodb"}}
	userRepo := db.New{{.Auth.UserCollection | title}}Repository(baseRepo.(*db.MongoRepository))
	{{end}}
	authSvc := authService.NewJWTAuthService(userRepo)
	userHdl := authHandler.NewUserHandler(authSvc, userRepo, "{{.Auth.UserCollection}}")
	{{else}}
	var authSvc authService.AuthService
	if os.Getenv("MOCK_AUTH") == "true" {
		log.Println("Using Mock Auth Service")
		authSvc = &authService.MockAuthService{}
	} else {
		// Initialize Firebase Auth
		app, err := firebase.NewApp(context.Background(), &firebase.Config{ProjectID: "{{.Database.ProjectID}}"})
		if err != nil {
			log.Fatalf("error initializing app: %v\n", err)
		}
		authClient, err := app.Auth(context.Background())
		if err != nil {
			log.Fatalf("error getting Auth client: %v\n", err)
		}
		authSvc = &authService.FirebaseAuthService{Client: authClient}
	}
	// Initialize User Handler
	{{if eq .Database.Type "firestore"}}
	userRepo := db.New{{.Auth.UserCollection | title}}Repository(baseRepo.(*db.FirestoreRepository))
	{{else if eq .Database.Type "postgresql"}}
	userRepo := db.New{{.Auth.UserCollection | title}}Repository(baseRepo.(*db.PostgresRepository))
	{{else if eq .Database.Type "mongodb"}}
	userRepo := db.New{{.Auth.UserCollection | title}}Repository(baseRepo.(*db.MongoRepository))
	{{end}}
	userHdl := authHandler.NewUserHandler(authSvc, userRepo, "{{.Auth.UserCollection}}")
	{{end}}
	{{end}}

	{{if and .Payments .Payments.Enabled}}
	// Initialize Payment Service
	{{if eq .Database.Type "firestore"}}
	mpRepo := db.New{{.Payments.TransactionsColl | title}}Repository(baseRepo.(*db.FirestoreRepository))
	{{else if eq .Database.Type "postgresql"}}
	mpRepo := db.New{{.Payments.TransactionsColl | title}}Repository(baseRepo.(*db.PostgresRepository))
	{{else if eq .Database.Type "mongodb"}}
	mpRepo := db.New{{.Payments.TransactionsColl | title}}Repository(baseRepo.(*db.MongoRepository))
	{{end}}

	{{if eq .Payments.Provider "mercadopago"}}
	mpService := payments.NewMercadoPagoService(mpRepo)
	{{else if eq .Payments.Provider "stripe"}}
	stripeService := payments.NewStripeService(mpRepo)
	{{end}}
	{{end}}

	// Setup Router
	r := gin.Default()

	// Swagger Route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))

	{{if and .Auth .Auth.Enabled}}
	// Auth Routes
	authGroup := r.Group("/auth")
	{{if eq .Auth.Provider "jwt"}}
	authGroup.POST("/login", userHdl.Login)
	authGroup.POST("/register", userHdl.Register)
	{{else}}
	authGroup.POST("/login", authService.AuthMiddleware(authSvc), userHdl.Login)
	{{end}}
	authGroup.GET("/me", authService.AuthMiddleware(authSvc), userHdl.GetMe)
	authGroup.GET("/roles", authService.AuthMiddleware(authSvc), userHdl.GetRoles)
	{{end}}

	{{if and .Payments .Payments.Enabled}}
	// Payment Routes
	paymentGroup := r.Group("/payments")
	{{if eq .Payments.Provider "mercadopago"}}
	paymentGroup.POST("/mercadopago/preference", mpService.CreatePreferenceHandler)
	paymentGroup.POST("/mercadopago/webhook", mpService.HandleWebhook)
	{{else if eq .Payments.Provider "stripe"}}
	paymentGroup.POST("/stripe/payment-intent", stripeService.CreatePaymentIntentHandler)
	paymentGroup.POST("/stripe/webhook", stripeService.HandleWebhook)
	{{end}}
	{{end}}

	{{range .Models}}
	// Routes for {{.Name}}
	{
		{{if eq $.Database.Type "firestore"}}
		repo := db.New{{.Name | title}}Repository(baseRepo.(*db.FirestoreRepository))
		{{else if eq $.Database.Type "postgresql"}}
		repo := db.New{{.Name | title}}Repository(baseRepo.(*db.PostgresRepository))
		{{else if eq $.Database.Type "mongodb"}}
		repo := db.New{{.Name | title}}Repository(baseRepo.(*db.MongoRepository))
		{{end}}
		handler := {{.Name | lower}}.New{{.Name | title}}Handler(repo)

		group := r.Group("/api/{{.Name}}")
		{{if .Protected}}
		{{if and $.Auth $.Auth.Enabled}}
		group.Use(authService.AuthMiddleware(authSvc))
		{{else}}
		group.Use(AuthMiddleware())
		{{end}}
		{{end}}
		group.GET("", handler.List)
		group.GET("/:id", handler.Get)
		group.POST("", handler.Create)
		group.PUT("/:id", handler.Update)
		group.DELETE("/:id", handler.Delete)
	}
	{{end}}

	log.Printf("Starting server for project: {{.ProjectName}}")
	r.Run(":8080")
}

{{if not (and .Auth .Auth.Enabled)}}
// AuthMiddleware verifies the Bearer token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}
		c.Next()
	}
}
{{end}}
`
	content, err := template.Render("main", mainTemplate, config)
	if err != nil {
		return err
	}
	return fs.WriteFile(filepath.Join(projectPath, "cmd/api/main.go"), content)
}
func generateTestScript(projectPath string, config *domain.Config, fs domain.FileSystemPort) error {
	var buf bytes.Buffer

	// Helper to generate JSON payload
	generateJSON := func(model domain.Model) string {
		var parts []string
		for k, v := range model.Fields {
			var val string
			switch v {
			case "string", "text":
				val = fmt.Sprintf("\"test_%s\"", k)
			case "integer", "int":
				val = "10"
			case "float":
				val = "99.99"
			case "boolean", "bool":
				val = "true"
			case "datetime":
				val = "\"2023-01-01T00:00:00Z\""
			default:
				val = "\"unknown\""
			}
			parts = append(parts, fmt.Sprintf("\"%s\": %s", k, val))
		}
		// Add relations
		for k, v := range model.Relations {
			if strings.HasPrefix(v, "belongsTo") {
				// If protected and relation is user_id, skip (injected by backend)
				if model.Protected && k == "user_id" {
					continue
				}
				parts = append(parts, fmt.Sprintf("\"%s\": \"test_%s\"", k, k))
			}
		}
		return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
	}

	buf.WriteString("#!/bin/bash\n\n")
	buf.WriteString("echo \"Installing dependencies...\"\n")
	buf.WriteString("go mod tidy\n\n")
	buf.WriteString("echo \"Generating docs...\"\n")
	buf.WriteString("./update_docs.sh\n\n")
	buf.WriteString("echo \"Starting server in background...\"\n")
	buf.WriteString("export MOCK_AUTH=true\n")
	if config.Payments != nil && config.Payments.Enabled {
		if config.Payments.Provider == "mercadopago" {
			buf.WriteString("export MP_ACCESS_TOKEN=\"TEST_MP_TOKEN_12345\"\n")
		} else if config.Payments.Provider == "stripe" {
			buf.WriteString("export STRIPE_SECRET_KEY=\"sk_test_12345\"\n")
			buf.WriteString("export STRIPE_WEBHOOK_SECRET=\"whsec_12345\"\n")
		}
	}
	buf.WriteString("go run cmd/api/main.go &\n")
	buf.WriteString("PID=$!\n")
	buf.WriteString("sleep 5\n\n")
	buf.WriteString("echo \"Running tests...\"\n\n")

	if config.Auth != nil && config.Auth.Enabled {
		if config.Auth.Provider == "jwt" {
			buf.WriteString("echo \"Testing POST /auth/register\"\n")
			buf.WriteString("curl -X POST -H \"Authorization: Bearer mock-token\" -H \"Content-Type: application/json\" -d '{\"email\": \"test@example.com\", \"password\": \"password123\"}' http://localhost:8080/auth/register\n")
			buf.WriteString("echo \"\\n\"\n")
		}

		buf.WriteString("echo \"Testing POST /auth/login\"\n")
		buf.WriteString("curl -X POST -H \"Authorization: Bearer mock-token\" -H \"Content-Type: application/json\" -d '{\"email\": \"test@example.com\", \"password\": \"password123\"}' http://localhost:8080/auth/login\n")
		buf.WriteString("echo \"\\n\"\n")
	}

	for _, model := range config.Models {
		payload := generateJSON(model)
		buf.WriteString(fmt.Sprintf("echo \"Testing POST /api/%s\"\n", model.Name))
		authHeader := ""
		if model.Protected {
			authHeader = "-H \"Authorization: Bearer mock-token\" "
		}
		buf.WriteString(fmt.Sprintf("curl -X POST %s-H \"Content-Type: application/json\" -d '%s' http://localhost:8080/api/%s\n", authHeader, payload, model.Name))
		buf.WriteString("echo \"\\n\"\n")
	}

	buf.WriteString("\necho \"Killing server (PID: $PID)...\"\n")
	buf.WriteString("kill $PID\n")

	if err := fs.WriteFile(filepath.Join(projectPath, "setup_and_test.sh"), buf.Bytes()); err != nil {
		return err
	}
	return fs.Chmod(filepath.Join(projectPath, "setup_and_test.sh"), 0755)
}

func generateSetupScript(projectPath string, config *domain.Config, fs domain.FileSystemPort) error {
	var buf bytes.Buffer
	buf.WriteString("#!/bin/bash\n\n")
	buf.WriteString("echo \"[1/3] Installing dependencies...\"\n")
	buf.WriteString("go mod tidy\n\n")
	buf.WriteString("export PATH=$PATH:$(go env GOPATH)/bin\n")
	buf.WriteString("if ! command -v swag &> /dev/null; then\n")
	buf.WriteString("    echo \"swag could not be found, installing...\"\n")
	buf.WriteString("    go install github.com/swaggo/swag/cmd/swag@latest\n")
	buf.WriteString("fi\n\n")
	buf.WriteString("echo \"[2/3] Generating docs...\"\n")
	buf.WriteString("./update_docs.sh\n\n")
	buf.WriteString("echo \"[3/3] Starting server...\"\n")
	buf.WriteString("echo \"Server will be available at http://localhost:8080\"\n")
	buf.WriteString("echo \"Swagger docs available at http://localhost:8080/swagger/index.html\"\n")
	if config.Payments != nil && config.Payments.Enabled {
		if config.Payments.Provider == "mercadopago" {
			buf.WriteString("export MP_ACCESS_TOKEN=\"YOUR_MERCADO_PAGO_ACCESS_TOKEN_HERE\"\n")
		} else if config.Payments.Provider == "stripe" {
			buf.WriteString("export STRIPE_SECRET_KEY=\"YOUR_STRIPE_SECRET_KEY_HERE\"\n")
			buf.WriteString("export STRIPE_WEBHOOK_SECRET=\"YOUR_STRIPE_WEBHOOK_SECRET_HERE\"\n")
		}
	}
	buf.WriteString("go run cmd/api/main.go\n")

	if err := fs.WriteFile(filepath.Join(projectPath, "setup.sh"), buf.Bytes()); err != nil {
		return err
	}
	return fs.Chmod(filepath.Join(projectPath, "setup.sh"), 0755)
}

func generateDocsScript(projectPath string, fs domain.FileSystemPort) error {
	var buf bytes.Buffer
	buf.WriteString("#!/bin/bash\n\n")
	buf.WriteString("export PATH=$PATH:$(go env GOPATH)/bin\n\n")
	buf.WriteString("if ! command -v swag &> /dev/null; then\n")
	buf.WriteString("    echo \"swag could not be found, installing...\"\n")
	buf.WriteString("    go install github.com/swaggo/swag/cmd/swag@latest\n")
	buf.WriteString("fi\n\n")
	buf.WriteString("echo \"Tidying dependencies...\"\n")
	buf.WriteString("go mod tidy\n\n")
	buf.WriteString("echo \"Generating Swagger documentation...\"\n")
	buf.WriteString("swag init -g cmd/api/main.go --parseDependency --parseInternal\n")

	if err := fs.WriteFile(filepath.Join(projectPath, "update_docs.sh"), buf.Bytes()); err != nil {
		return err
	}
	return fs.Chmod(filepath.Join(projectPath, "update_docs.sh"), 0755)
}

func generateModelHandlerTests(projectPath string, config *domain.Config, model domain.Model, fs domain.FileSystemPort, template domain.TemplatePort) error {
	const testTemplate = `package {{.Model.Name | lower}}

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"{{.ProjectName}}/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type Mock{{.Model.Name | title}}Repository struct {
	Data map[string]*domain.{{.Model.Name | title}}
}

func (m *Mock{{.Model.Name | title}}Repository) List(ctx context.Context, limit, offset int) ([]*domain.{{.Model.Name | title}}, error) {
	var results []*domain.{{.Model.Name | title}}
	for _, v := range m.Data {
		results = append(results, v)
	}
	
	// Simple slicing for mock pagination
	if offset >= len(results) {
		return []*domain.{{.Model.Name | title}}{}, nil
	}
	end := offset + limit
	if end > len(results) {
		end = len(results)
	}
	return results[offset:end], nil
}

func (m *Mock{{.Model.Name | title}}Repository) Get(ctx context.Context, id string) (*domain.{{.Model.Name | title}}, error) {
	if val, ok := m.Data[id]; ok {
		return val, nil
	}
	return nil, nil
}

func (m *Mock{{.Model.Name | title}}Repository) Create(ctx context.Context, model *domain.{{.Model.Name | title}}) (string, error) {
	id := "test-id"
	model.ID = id
	if m.Data == nil {
		m.Data = make(map[string]*domain.{{.Model.Name | title}})
	}
	m.Data[id] = model
	return id, nil
}

func (m *Mock{{.Model.Name | title}}Repository) Update(ctx context.Context, id string, model *domain.{{.Model.Name | title}}) error {
	m.Data[id] = model
	return nil
}

func (m *Mock{{.Model.Name | title}}Repository) Delete(ctx context.Context, id string) error {
	delete(m.Data, id)
	return nil
}

func Test{{.Model.Name | title}}Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &Mock{{.Model.Name | title}}Repository{Data: make(map[string]*domain.{{.Model.Name | title}})}
	handler := New{{.Model.Name | title}}Handler(repo)
	r := gin.Default()

	r.GET("/{{.Model.Name | lower}}", handler.List)
	r.POST("/{{.Model.Name | lower}}", handler.Create)

	t.Run("Create", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := domain.{{.Model.Name | title}}{}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/{{.Model.Name | lower}}", bytes.NewBuffer(jsonBody))
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("List", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/{{.Model.Name | lower}}?page=1&limit=10", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
`
	data := struct {
		ProjectName string
		Model       domain.Model
	}{
		ProjectName: config.ProjectName,
		Model:       model,
	}

	content, err := template.Render(model.Name+"_test", testTemplate, data)
	if err != nil {
		return err
	}
	return fs.WriteFile(filepath.Join(projectPath, "internal/handlers", strings.ToLower(model.Name), "handler_test.go"), content)
}

func generateArchitectureDocs(projectPath string, fs domain.FileSystemPort) error {
	const content = `# Architecture of Generated Project

This project follows **Hexagonal Architecture** (also known as Ports and Adapters) and **Clean Code** principles. It supports multiple database drivers (Firestore, PostgreSQL, MongoDB) through a unified repository interface.

## Directory Structure

` + "```" + `
<project_name>/
├── cmd/
│   └── api/
│       └── main.go           # Entry point: wires everything together
├── internal/
│   ├── domain/               # Core business logic (Ports)
│   │   ├── <model>.go        # Model struct and Repository interface
│   ├── infrastructure/       # External concerns (Adapters)
│   │   └── db/
│   │       ├── firestore.go  # Firestore client (if selected)
│   │       ├── postgres.go   # PostgreSQL client (if selected)
│   │       ├── mongo.go      # MongoDB client (if selected)
│   │       └── <model>_repo.go # DB-specific implementation of the Port
│   ├── handlers/             # Application Layer (Adapters)
│   │   ├── <model>/
│   │   │   ├── handler.go    # HTTP handlers for the model
│   │   │   └── handler_test.go # Unit tests for the handler
│   │   └── auth/             # Authentication handlers
│   ├── auth/                 # Auth logic and middleware
│   ├── payments/             # Payment provider integrations
│   └── config/               # Configuration management
└── ...
` + "```" + `

## Core Concepts

### 1. Database Abstraction

The project uses a **Repository Pattern** to abstract database operations. The domain layer defines interfaces (Ports), and the infrastructure layer provides implementations (Adapters) for the selected database:

- **Firestore**: Uses the official Google Cloud Firestore SDK.
- **PostgreSQL**: Uses ` + "`" + `pgx` + "`" + ` for high-performance SQL operations.
- **MongoDB**: Uses the official MongoDB Go driver.

### 2. Dependency Injection

All dependencies are injected in ` + "`" + `cmd/api/main.go` + "`" + `. The database client is initialized based on the configuration and passed to the model-specific repositories.

### 3. Clean Code & Guard Clauses

The code uses **guard clauses** to keep the logic flat and readable, avoiding deep nesting.

## How to Work with This Code

1.  **Switching Databases**: To change the database, you would typically update the initialization in ` + "`" + `main.go` + "`" + ` and provide the corresponding repository implementation in ` + "`" + `internal/infrastructure/db` + "`" + `.
2.  **Adding a Field**: Update the domain struct and the repository implementation.
3.  **Environment Variables**:
    - ` + "`" + `DATABASE_URL` + "`" + `: Required for PostgreSQL and MongoDB.
    - ` + "`" + `MOCK_AUTH` + "`" + `: Set to ` + "`" + `true` + "`" + ` to bypass Firebase Auth during development.
`
	return fs.WriteFile(filepath.Join(projectPath, "ARCHITECTURE.md"), []byte(content))
}
