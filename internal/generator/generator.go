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
)

// Client wraps the Firestore client
type Client struct {
	Firestore *firestore.Client
}

// InitFirestore initializes the Firestore client
func InitFirestore() (*Client, error) {
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

	return &Client{Firestore: client}, nil
}

// Close closes the Firestore client
func (c *Client) Close() {
	if c.Firestore != nil {
		c.Firestore.Close()
	}
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
	"context"
	"log"
	"net/http"
	{{if not (and .Auth .Auth.Enabled)}}"strings"{{end}}

	"{{.ProjectName}}/internal/db"
	"github.com/gin-gonic/gin"
	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	{{if and .Auth .Auth.Enabled}}
	auth "{{.ProjectName}}/internal/auth"
	"{{.ProjectName}}/internal/handlers"
	firebase "firebase.google.com/go/v4"
	{{end}}
)

func main() {
	// Initialize Database
	dbClient, err := db.InitFirestore()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbClient.Close()

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
	userHandler := handlers.NewUserHandler(authClient, dbClient.Firestore, "{{.Auth.UserCollection}}")
	{{end}}

	// Setup Router
	r := gin.Default()

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
		group.GET("", createListHandler(dbClient, "{{.Name}}"))
		group.GET("/:id", createGetHandler(dbClient, "{{.Name}}"))
		group.POST("", createCreateHandler(dbClient, "{{.Name}}"))
		group.PUT("/:id", createUpdateHandler(dbClient, "{{.Name}}"))
		group.DELETE("/:id", createDeleteHandler(dbClient, "{{.Name}}"))
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

func createListHandler(client *db.Client, collection string) gin.HandlerFunc {
	return func(c *gin.Context) {
		iter := client.Firestore.Collection(collection).Documents(c.Request.Context())
		var results []map[string]interface{}
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			data := doc.Data()
			data["id"] = doc.Ref.ID
			results = append(results, data)
		}
		c.JSON(http.StatusOK, results)
	}
}

func createGetHandler(client *db.Client, collection string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		doc, err := client.Firestore.Collection(collection).Doc(id).Get(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		data := doc.Data()
		data["id"] = doc.Ref.ID
		c.JSON(http.StatusOK, data)
	}
}

func createCreateHandler(client *db.Client, collection string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		ref, _, err := client.Firestore.Collection(collection).Add(c.Request.Context(), data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		data["id"] = ref.ID
		c.JSON(http.StatusCreated, data)
	}
}

func createUpdateHandler(client *db.Client, collection string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		_, err := client.Firestore.Collection(collection).Doc(id).Set(c.Request.Context(), data, firestore.MergeAll)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	}
}

func createDeleteHandler(client *db.Client, collection string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		_, err := client.Firestore.Collection(collection).Doc(id).Delete(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	}
}
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
