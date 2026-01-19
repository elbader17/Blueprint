package main

import (
	"context"
	"log"
	"os"
	

	"github.com/joho/godotenv"
	"test/internal/infrastructure/db"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "test/docs"
	
	authService "test/internal/auth"
	authHandler "test/internal/handlers/auth"
	firebase "firebase.google.com/go/v4"
	
	
	
	"test/internal/handlers/account"
	
	"test/internal/handlers/user"
	
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize Database
	
	baseRepo, err := db.NewFirestoreRepository()
	
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer baseRepo.Close()

	
	// Initialize Auth Service
	var authSvc authService.AuthService
	if os.Getenv("MOCK_AUTH") == "true" {
		log.Println("Using Mock Auth Service")
		authSvc = &authService.MockAuthService{}
	} else {
		// Initialize Firebase Auth
		app, err := firebase.NewApp(context.Background(), &firebase.Config{ProjectID: "your-project-id"})
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
	
	userRepo := db.NewUserRepository(baseRepo.(*db.FirestoreRepository))
	
	userHdl := authHandler.NewUserHandler(authSvc, userRepo, "User")
	

	

	// Setup Router
	r := gin.Default()

	// Swagger Route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))

	
	// Auth Routes
	authGroup := r.Group("/auth")
	authGroup.POST("/login", authService.AuthMiddleware(authSvc), userHdl.Login)
	authGroup.POST("/register", authService.AuthMiddleware(authSvc), userHdl.Login)
	authGroup.GET("/me", authService.AuthMiddleware(authSvc), userHdl.GetMe)
	authGroup.GET("/roles", authService.AuthMiddleware(authSvc), userHdl.GetRoles)
	

	

	
	// Routes for account
	{
		
		repo := db.NewAccountRepository(baseRepo.(*db.FirestoreRepository))
		
		handler := account.NewAccountHandler(repo)

		group := r.Group("/api/account")
		
		
		group.Use(authService.AuthMiddleware(authSvc))
		
		
		group.GET("", handler.List)
		group.GET("/:id", handler.Get)
		group.POST("", handler.Create)
		group.PUT("/:id", handler.Update)
		group.DELETE("/:id", handler.Delete)
	}
	
	// Routes for User
	{
		
		repo := db.NewUserRepository(baseRepo.(*db.FirestoreRepository))
		
		handler := user.NewUserHandler(repo)

		group := r.Group("/api/User")
		
		
		group.Use(authService.AuthMiddleware(authSvc))
		
		
		group.GET("", handler.List)
		group.GET("/:id", handler.Get)
		group.POST("", handler.Create)
		group.PUT("/:id", handler.Update)
		group.DELETE("/:id", handler.Delete)
	}
	

	log.Printf("Starting server for project: test")
	r.Run(":8080")
}


