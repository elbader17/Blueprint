package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/eduardo/blueprint/internal/parser"
)

// Generate creates the API project based on the config
func Generate(config *parser.Config, outputDir string) error {
	projectPath := filepath.Join(outputDir, config.ProjectName)
	fmt.Printf("Creating project at %s\n", projectPath)
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create directories
	dirs := []string{
		"cmd/api",
		"internal/db",
		"internal/auth",
		"internal/handlers",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(projectPath, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate go.mod
	if err := generateGoMod(projectPath, config.ProjectName); err != nil {
		return err
	}

	// Generate internal/db/firestore.go
	if err := generateFirestore(projectPath, config); err != nil {
		return err
	}

	// Generate Auth files if enabled
	if config.Auth != nil && config.Auth.Enabled {
		if err := generateAuthFiles(projectPath); err != nil {
			return err
		}
	}

	// Copy firebaseCredentials.json
	if err := copyFile("firebaseCredentials.json", filepath.Join(projectPath, "firebaseCredentials.json")); err != nil {
		// Warn but don't fail if credentials don't exist, maybe user wants to add them later
		fmt.Printf("Warning: firebaseCredentials.json not found or could not be copied: %v\n", err)
	}

	// Generate cmd/api/main.go
	if err := generateMain(projectPath, config); err != nil {
		return err
	}

	// Generate setup_and_test.sh
	if err := generateTestScript(projectPath, config); err != nil {
		return err
	}

	// Generate update_docs.sh
	if err := generateDocsScript(projectPath); err != nil {
		return err
	}

	// Generate tests
	if err := generateHandlerTests(projectPath, config); err != nil {
		return err
	}

	return nil
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}

func generateGoMod(projectPath, projectName string) error {
	content := fmt.Sprintf(`module %s

go 1.21

require (
	cloud.google.com/go/firestore v1.14.0
	firebase.google.com/go/v4 v4.13.0
	github.com/gin-gonic/gin v1.9.1
	google.golang.org/api v0.150.0
	github.com/swaggo/files v1.0.1
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/swag v1.16.2
)
`, projectName)
	return os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte(content), 0644)
}

func generateFirestore(projectPath string, config *parser.Config) error {
	// Embedding the content of firestore.go directly for simplicity in this MVP
	// We use the ProjectID from config, defaulting to a placeholder if empty
	projectID := config.FirestoreProjectID
	if projectID == "" {
		projectID = "tiendaonline-mvp"
	}

	content := fmt.Sprintf(`package db

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
	conf := &firebase.Config{ProjectID: "%s"}
	
	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %%v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing firestore: %%v", err)
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
`, projectID)
	return os.WriteFile(filepath.Join(projectPath, "internal/db/firestore.go"), []byte(content), 0644)
}

func generateAuthFiles(projectPath string) error {
	if err := os.WriteFile(filepath.Join(projectPath, "internal/auth/middleware.go"), []byte(AuthMiddlewareTemplate), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(projectPath, "internal/handlers/auth.go"), []byte(AuthHandlerTemplate), 0644); err != nil {
		return err
	}
	return nil
}

func generateMain(projectPath string, config *parser.Config) error {
	const mainTemplate = `package main

import (
	{{if and .Auth .Auth.Enabled}}"context"{{end}}
	"log"
	"net/http"
	{{if not (and .Auth .Auth.Enabled)}}"strings"{{end}}

	"{{.ProjectName}}/internal/db"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "{{.ProjectName}}/docs"
	{{if and .Auth .Auth.Enabled}}
	auth "{{.ProjectName}}/internal/auth"
	"{{.ProjectName}}/internal/handlers"
	firebase "firebase.google.com/go/v4"
	{{end}}
)

var repo db.Repository

// @title {{.ProjectName}} API
// @version 1.0
// @description API generated by Blueprint
// @host localhost:8080
// @BasePath /
func main() {
	var err error
	// Initialize Database
	repo, err = db.NewFirestoreRepository()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer repo.Close()

	{{if and .Auth .Auth.Enabled}}
	// Initialize Firebase Auth
	app, err := firebase.NewApp(context.Background(), &firebase.Config{ProjectID: "{{.FirestoreProjectID}}"})
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	// Initialize User Handler
	userHandler := handlers.NewUserHandler(authClient, repo.GetClient(), "{{.Auth.UserCollection}}")
	{{end}}

	// Setup Router
	r := gin.Default()

	// Swagger Route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))

	{{if and .Auth .Auth.Enabled}}
	// Auth Routes
	authGroup := r.Group("/auth")
	authGroup.POST("/login", auth.AuthMiddleware(authClient), userHandler.Login)
	authGroup.POST("/register", auth.AuthMiddleware(authClient), userHandler.Login)
	authGroup.GET("/me", auth.AuthMiddleware(authClient), userHandler.GetMe)
	authGroup.GET("/roles", auth.AuthMiddleware(authClient), userHandler.GetRoles)
	{{end}}

	{{range .Models}}
	// Routes for {{.Name}}
	{
		group := r.Group("/api/{{.Name}}")
		{{if .Protected}}
		{{if and $.Auth $.Auth.Enabled}}
		group.Use(auth.AuthMiddleware(authClient))
		{{else}}
		group.Use(AuthMiddleware())
		{{end}}
		{{end}}
		group.GET("", list{{.Name}}Handler)
		group.GET("/:id", get{{.Name}}Handler)
		group.POST("", create{{.Name}}Handler)
		group.PUT("/:id", update{{.Name}}Handler)
		group.DELETE("/:id", delete{{.Name}}Handler)
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

// Generic Handlers

func genericListHandler(collection string) gin.HandlerFunc {
	return func(c *gin.Context) {
		results, err := repo.List(c.Request.Context(), collection)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, results)
	}
}

func genericGetHandler(collection string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		data, err := repo.Get(c.Request.Context(), collection, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	}
}

func genericCreateHandler(collection string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		id, err := repo.Create(c.Request.Context(), collection, data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		data["id"] = id
		c.JSON(http.StatusCreated, data)
	}
}

func genericUpdateHandler(collection string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		if err := repo.Update(c.Request.Context(), collection, id, data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	}
}

func genericDeleteHandler(collection string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := repo.Delete(c.Request.Context(), collection, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	}
}

// Specific Handlers
{{range .Models}}
// list{{.Name}}Handler godoc
// @Summary List {{.Name}}
// @Description Get all {{.Name}}
// @Tags {{.Name}}
// @Accept  json
// @Produce  json
// @Success 200 {array} map[string]interface{}
// @Router /api/{{.Name}} [get]
func list{{.Name}}Handler(c *gin.Context) {
	genericListHandler("{{.Name}}")(c)
}

// get{{.Name}}Handler godoc
// @Summary Get {{.Name}}
// @Description Get a {{.Name}} by ID
// @Tags {{.Name}}
// @Accept  json
// @Produce  json
// @Param id path string true "{{.Name}} ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/{{.Name}}/{id} [get]
func get{{.Name}}Handler(c *gin.Context) {
	genericGetHandler("{{.Name}}")(c)
}

// create{{.Name}}Handler godoc
// @Summary Create {{.Name}}
// @Description Create a new {{.Name}}
// @Tags {{.Name}}
// @Accept  json
// @Produce  json
// @Param {{.Name}} body map[string]interface{} true "New {{.Name}}"
// @Success 201 {object} map[string]interface{}
// @Router /api/{{.Name}} [post]
func create{{.Name}}Handler(c *gin.Context) {
	genericCreateHandler("{{.Name}}")(c)
}

// update{{.Name}}Handler godoc
// @Summary Update {{.Name}}
// @Description Update a {{.Name}} by ID
// @Tags {{.Name}}
// @Accept  json
// @Produce  json
// @Param id path string true "{{.Name}} ID"
// @Param {{.Name}} body map[string]interface{} true "Updated {{.Name}}"
// @Success 200 {object} map[string]interface{}
// @Router /api/{{.Name}}/{id} [put]
func update{{.Name}}Handler(c *gin.Context) {
	genericUpdateHandler("{{.Name}}")(c)
}

// delete{{.Name}}Handler godoc
// @Summary Delete {{.Name}}
// @Description Delete a {{.Name}} by ID
// @Tags {{.Name}}
// @Accept  json
// @Produce  json
// @Param id path string true "{{.Name}} ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/{{.Name}}/{id} [delete]
func delete{{.Name}}Handler(c *gin.Context) {
	genericDeleteHandler("{{.Name}}")(c)
}
{{end}}
`
	tmpl, err := template.New("main").Parse(mainTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(projectPath, "cmd/api/main.go"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, config)
}

func generateTestScript(projectPath string, config *parser.Config) error {
	scriptPath := filepath.Join(projectPath, "setup_and_test.sh")
	f, err := os.Create(scriptPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Make script executable
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return err
	}

	// Helper to generate JSON payload
	generateJSON := func(fields map[string]string) string {
		var parts []string
		for k, v := range fields {
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
		return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
	}

	// Write script content
	fmt.Fprintf(f, "#!/bin/bash\n\n")
	fmt.Fprintf(f, "echo \"Installing dependencies...\"\n")
	fmt.Fprintf(f, "go mod tidy\n\n")

	fmt.Fprintf(f, "echo \"Generating docs...\"\n")
	fmt.Fprintf(f, "./update_docs.sh\n\n")

	fmt.Fprintf(f, "echo \"Starting server in background...\"\n")
	fmt.Fprintf(f, "go run cmd/api/main.go &\n")
	fmt.Fprintf(f, "PID=$!\n")
	fmt.Fprintf(f, "sleep 5\n\n") // Wait for server to start

	fmt.Fprintf(f, "echo \"Running tests...\"\n\n")

	for _, model := range config.Models {
		payload := generateJSON(model.Fields)
		fmt.Fprintf(f, "echo \"Testing POST /api/%s\"\n", model.Name)
		
		authHeader := ""
		if model.Protected {
			authHeader = "-H \"Authorization: Bearer mock-token\" "
		}

		fmt.Fprintf(f, "curl -X POST %s-H \"Content-Type: application/json\" -d '%s' http://localhost:8080/api/%s\n", authHeader, payload, model.Name)
		fmt.Fprintf(f, "echo \"\\n\"\n")
	}

	fmt.Fprintf(f, "\necho \"Killing server (PID: $PID)...\"\n")
	fmt.Fprintf(f, "kill $PID\n")
	
	return nil
}

func generateDocsScript(projectPath string) error {
	scriptPath := filepath.Join(projectPath, "update_docs.sh")
	f, err := os.Create(scriptPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := os.Chmod(scriptPath, 0755); err != nil {
		return err
	}

	fmt.Fprintf(f, "#!/bin/bash\n\n")
	fmt.Fprintf(f, "echo \"Generating Swagger documentation...\"\n")
	fmt.Fprintf(f, "swag init -g cmd/api/main.go\n")
	
	return nil
}

func generateHandlerTests(projectPath string, config *parser.Config) error {
	const testTemplate = `package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"{{.ProjectName}}/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"cloud.google.com/go/firestore"
)

// MockRepository implements db.Repository for testing
type MockRepository struct {
	Data map[string]map[string]map[string]interface{}
}

func NewMockRepository() db.Repository {
	return &MockRepository{
		Data: make(map[string]map[string]map[string]interface{}),
	}
}

func (m *MockRepository) Close() {}
func (m *MockRepository) GetClient() *firestore.Client { return nil }

func (m *MockRepository) List(ctx context.Context, collection string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	if cols, ok := m.Data[collection]; ok {
		for _, v := range cols {
			results = append(results, v)
		}
	}
	return results, nil
}

func (m *MockRepository) Get(ctx context.Context, collection, id string) (map[string]interface{}, error) {
	if cols, ok := m.Data[collection]; ok {
		if val, ok := cols[id]; ok {
			return val, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockRepository) Create(ctx context.Context, collection string, data map[string]interface{}) (string, error) {
	if m.Data[collection] == nil {
		m.Data[collection] = make(map[string]map[string]interface{})
	}
	id := "test-id"
	data["id"] = id
	m.Data[collection][id] = data
	return id, nil
}

func (m *MockRepository) Update(ctx context.Context, collection, id string, data map[string]interface{}) error {
	if m.Data[collection] == nil {
		return fmt.Errorf("not found")
	}
	if _, ok := m.Data[collection][id]; !ok {
		return fmt.Errorf("not found")
	}
	// Merge
	for k, v := range data {
		m.Data[collection][id][k] = v
	}
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, collection, id string) error {
	if m.Data[collection] != nil {
		delete(m.Data[collection], id)
	}
	return nil
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	return r
}

{{range .Models}}
func Test{{.Name | title}}Handlers(t *testing.T) {
	// Setup
	mockRepo := NewMockRepository()
	repo = mockRepo // Inject mock
	r := setupTestRouter()
	
	// Register routes
	group := r.Group("/api/{{.Name}}")
	group.GET("", list{{.Name}}Handler)
	group.GET("/:id", get{{.Name}}Handler)
	group.POST("", create{{.Name}}Handler)
	group.PUT("/:id", update{{.Name}}Handler)
	group.DELETE("/:id", delete{{.Name}}Handler)

	t.Run("Create {{.Name}}", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := map[string]interface{}{
			"test_field": "test_value",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/{{.Name}}", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("List {{.Name}}", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/{{.Name}}", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
{{end}}
`
	funcMap := template.FuncMap{
		"title": func(s string) string {
			if s == "" {
				return ""
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
	}

	tmpl, err := template.New("test").Funcs(funcMap).Parse(testTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(projectPath, "cmd/api/handlers_test.go"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, config)
}
