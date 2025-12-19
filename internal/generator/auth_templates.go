package generator

const AuthMiddlewareTemplate = `package auth

import (
	"context"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

// AuthService defines the interface for authentication
type AuthService interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

// FirebaseAuthService implements AuthService using Firebase
type FirebaseAuthService struct {
	Client *auth.Client
}

func (s *FirebaseAuthService) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	return s.Client.VerifyIDToken(ctx, idToken)
}

// MockAuthService implements AuthService for testing
type MockAuthService struct {}

func (m *MockAuthService) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	// Return a valid mock token
	return &auth.Token{
		UID: "test-user-id",
		Claims: map[string]interface{}{
			"email": "test@example.com",
			"name": "Test User",
		},
	}, nil
}

// AuthMiddleware verifies the Firebase ID token
func AuthMiddleware(service AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing Authorization header",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Authorization header format",
			})
			return
		}

		token, err := service.VerifyIDToken(context.Background(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}

		// Store user info in context
		c.Set("user", token)
		c.Next()
	}
}
`

const AuthHandlerTemplate = `package auth

import (
	"context"
	"log"
	"net/http"

	"{{.ProjectName}}/internal/auth"
	"{{.ProjectName}}/internal/domain"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	AuthService    auth.AuthService
	Repository     domain.{{.Auth.UserCollection | title}}Repository
	UserCollection string
}

func NewUserHandler(authService auth.AuthService, repo domain.{{.Auth.UserCollection | title}}Repository, userCollection string) *UserHandler {
	return &UserHandler{
		AuthService:    authService,
		Repository:     repo,
		UserCollection: userCollection,
	}
}

// Login godoc
// @Summary Login or Register
// @Description Login with Firebase token and sync user data
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param body body object{role=string,settings=object} false "Optional role and settings"
// @Success 200 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	type LoginRequest struct {
		Role     string                 ` + "`" + `json:"role"` + "`" + `
		Settings map[string]interface{} ` + "`" + `json:"settings"` + "`" + `
	}
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Continue even if body is invalid, as it's optional
	}

	userTokenInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	userToken := userTokenInterface.(*firebaseAuth.Token)
	uid := userToken.UID
	
	var email string
	if e, ok := userToken.Claims["email"].(string); ok {
		email = e
	}

	// Check if user exists using Repository
	docSnap, err := h.Repository.Get(context.Background(), uid)
	
	isNewUser := (err != nil || docSnap == nil)

	data := &domain.{{.Auth.UserCollection | title}}{
		ID: uid,
	}
	if email != "" {
		data.Email = email
	}

	if name, ok := userToken.Claims["name"].(string); ok {
		data.Name = name
	}
	if picture, ok := userToken.Claims["picture"].(string); ok {
		data.Picture = picture
	}

	if isNewUser {
		roleId := "admin"
		if req.Role != "" {
			roleId = req.Role
		}
		data.RoleId = roleId

		// Note: Simplified settings creation for this template
		// In a real scenario, you'd have a SettingsRepository
	}

	if !isNewUser && req.Role != "" {
		data.RoleId = req.Role
	}

	// Use Update for Upsert behavior
	err = h.Repository.Update(context.Background(), uid, data)
	if err != nil {
		log.Printf("Failed to update user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User logged in and synced",
		"uid":     uid,
		"email":   email,
		"roleId":  data.RoleId,
	})
}

// GetMe godoc
// @Summary Get Current User
// @Description Get profile of the authenticated user
// @Tags Auth
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Router /auth/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userTokenInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	userToken := userTokenInterface.(*firebaseAuth.Token)
	uid := userToken.UID

	userData, err := h.Repository.Get(context.Background(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, userData)
}

// GetRoles godoc
// @Summary List Roles
// @Description Get available roles
// @Tags Auth
// @Accept  json
// @Produce  json
// @Success 200 {array} map[string]interface{}
// @Router /auth/roles [get]
func (h *UserHandler) GetRoles(c *gin.Context) {
	// This would typically use a RoleRepository
	c.JSON(http.StatusOK, []string{"admin", "user"})
}
`

