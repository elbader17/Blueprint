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

const JWTMiddlewareTemplate = `package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"{{.ProjectName}}/internal/config"
	"{{.ProjectName}}/internal/domain"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines the interface for authentication
type AuthService interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
	Login(ctx context.Context, email, password string) (string, error)
	Register(ctx context.Context, email, password string) (string, error)
}

// CustomClaims for JWT
type CustomClaims struct {
	UserID string ` + "`json:\"uid\"`" + `
	Email  string ` + "`json:\"email\"`" + `
	Role   string ` + "`json:\"role\"`" + `
	jwt.RegisteredClaims
}

// JWTAuthService implements AuthService using JWT
type JWTAuthService struct {
	SecretKey []byte
	Repo      domain.UserRepository
}

func NewJWTAuthService(repo domain.UserRepository) *JWTAuthService {
	return &JWTAuthService{
		SecretKey: []byte(config.GetJWTSecret()),
		Repo:      repo,
	}
}

func (s *JWTAuthService) VerifyIDToken(ctx context.Context, tokenString string) (*auth.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.SecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		// Map JWT claims to Firebase-like auth.Token for compatibility
		return &auth.Token{
			UID: claims.UserID,
			Claims: map[string]interface{}{
				"email": claims.Email,
				"role":  claims.Role,
			},
		}, nil
	}

	return nil, errors.New("invalid token")
}

func (s *JWTAuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.Repo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	claims := CustomClaims{
		user.ID,
		user.Email,
		user.Role,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.SecretKey)
}

func (s *JWTAuthService) Register(ctx context.Context, email, password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &domain.UserAuthData{
		Email:    email,
		Password: string(hashedPassword),
		Role:     "user", // Default role
	}
	
	id, err := s.Repo.RegisterUser(ctx, user)
	if err != nil {
		return "", err
	}
	return id, nil
}

// AuthMiddleware verifies the JWT token
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

const JWTAuthHandlerTemplate = `package auth

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
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Email    string ` + "`" + `json:"email" binding:"required"` + "`" + `
		Password string ` + "`" + `json:"password" binding:"required"` + "`" + `
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.AuthService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Register godoc
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Email    string ` + "`" + `json:"email" binding:"required"` + "`" + `
		Password string ` + "`" + `json:"password" binding:"required"` + "`" + `
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.AuthService.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "User created successfully"})
}

// GetMe godoc
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

	// Don't leak the password hash
	if u, ok := interface{}(userData).(*domain.{{.Auth.UserCollection | title}}); ok {
		u.Password = ""
	}

	c.JSON(http.StatusOK, userData)
}

// GetRoles godoc
func (h *UserHandler) GetRoles(c *gin.Context) {
	c.JSON(http.StatusOK, []string{"admin", "user"})
}
`
