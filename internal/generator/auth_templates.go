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

const AuthHandlerTemplate = `package handlers

import (
	"context"
	"log"
	"net/http"

	"{{.ProjectName}}/internal/auth"
	"{{.ProjectName}}/internal/db"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	AuthService    auth.AuthService
	Repository     db.Repository
	UserCollection string
}

func NewUserHandler(authService auth.AuthService, repo db.Repository, userCollection string) *UserHandler {
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
	if email == "" {
		log.Printf("Warning: No email found in token claims for UID: %s", uid)
	}

	// Check if user exists using Repository
	docSnap, err := h.Repository.Get(context.Background(), h.UserCollection, uid)
	
	isNewUser := false
	if err != nil {
		// Assuming error means not found or other issue, treat as new
		isNewUser = true
	} else if docSnap == nil {
		isNewUser = true
	}

	data := map[string]interface{}{
		"uid": uid,
	}
	if email != "" {
		data["email"] = email
	}

	if name, ok := userToken.Claims["name"].(string); ok {
		data["name"] = name
	}
	if picture, ok := userToken.Claims["picture"].(string); ok {
		data["picture"] = picture
	}

	if isNewUser {
		roleId := "admin"
		if req.Role != "" {
			roleId = req.Role
		}

		if roleId == "admin" {
			// Ensure admin role exists
			err := h.Repository.Update(context.Background(), "roles", "admin", map[string]interface{}{
				"name": "Admin",
			})
			if err != nil {
				// Try create if update failed (likely not found)
				// Note: Repository.Update usually implies existence check in some implementations, 
				// but here we use it as upsert if possible or fallback.
				// Actually, our FirestoreRepository.Update uses Set with MergeAll, so it acts as Upsert.
				log.Printf("Ensured admin role: %v", err)
			}
		}
		
		data["roleId"] = roleId

		settingsData := map[string]interface{}{
			"test": "test",
		}
		if req.Settings != nil {
			for k, v := range req.Settings {
				settingsData[k] = v
			}
		}

		settingsId, err := h.Repository.Create(context.Background(), "settings", settingsData)
		if err != nil {
			log.Printf("Failed to create settings doc: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create settings"})
			return
		}

		data["settingsId"] = settingsId
	}

	if !isNewUser && req.Role != "" {
		data["roleId"] = req.Role
	}

	log.Printf("Attempting to save user data for UID %s (New: %v): %+v", uid, isNewUser, data)

	// Use Update for Upsert behavior (Set with MergeAll)
	err = h.Repository.Update(context.Background(), h.UserCollection, uid, data)

	if err != nil {
		log.Printf("Failed to update user in Firestore: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User logged in and synced",
		"uid":     uid,
		"email":   email,
		"roleId":  data["roleId"],
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

	userData, err := h.Repository.Get(context.Background(), h.UserCollection, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if roleId, ok := userData["roleId"].(string); ok {
		roleData, err := h.Repository.Get(context.Background(), "roles", roleId)
		if err == nil {
			userData["role"] = roleData
		}
	}

	if settingsId, ok := userData["settingsId"].(string); ok {
		settingsData, err := h.Repository.Get(context.Background(), "settings", settingsId)
		if err == nil {
			userData["settings"] = settingsData
		}
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
	roles, err := h.Repository.List(context.Background(), "roles")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch roles"})
		return
	}
	c.JSON(http.StatusOK, roles)
}
`

