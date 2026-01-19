package auth

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
