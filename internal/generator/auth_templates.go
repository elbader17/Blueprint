package generator

const AuthMiddlewareTemplate = `package middleware

import (
	"context"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifies the Firebase ID token
func AuthMiddleware(client *auth.Client) gin.HandlerFunc {
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

		token, err := client.VerifyIDToken(context.Background(), tokenString)
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

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

type UserHandler struct {
	AuthClient      *auth.Client
	FirestoreClient *firestore.Client
	UserCollection  string
}

func NewUserHandler(auth *auth.Client, fs *firestore.Client, userCollection string) *UserHandler {
	return &UserHandler{
		AuthClient:      auth,
		FirestoreClient: fs,
		UserCollection:  userCollection,
	}
}

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
	userToken := userTokenInterface.(*auth.Token)
	uid := userToken.UID
	
	var email string
	if e, ok := userToken.Claims["email"].(string); ok {
		email = e
	}
	if email == "" {
		log.Printf("Warning: No email found in token claims for UID: %s", uid)
	}

	docRef := h.FirestoreClient.Collection(h.UserCollection).Doc(uid)
	docSnap, err := docRef.Get(context.Background())
	
	isNewUser := false
	if err != nil {
		isNewUser = true
	}
	if err == nil && !docSnap.Exists() {
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
			_, err := h.FirestoreClient.Collection("roles").Doc("admin").Set(context.Background(), map[string]interface{}{
				"name": "Admin",
			}, firestore.MergeAll)
			if err != nil {
				log.Printf("Failed to ensure admin role exists: %v", err)
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

		newSettingsRef, _, err := h.FirestoreClient.Collection("settings").Add(context.Background(), settingsData)
		if err != nil {
			log.Printf("Failed to create settings doc: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create settings"})
			return
		}

		data["settingsId"] = newSettingsRef.ID
	}

	if !isNewUser && req.Role != "" {
		data["roleId"] = req.Role
	}

	log.Printf("Attempting to save user data for UID %s (New: %v): %+v", uid, isNewUser, data)

	_, err = docRef.Set(context.Background(), data, firestore.MergeAll)

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

func (h *UserHandler) GetMe(c *gin.Context) {
	userTokenInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	userToken := userTokenInterface.(*auth.Token)
	uid := userToken.UID

	doc, err := h.FirestoreClient.Collection(h.UserCollection).Doc(uid).Get(context.Background())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	userData := doc.Data()

	if roleId, ok := userData["roleId"].(string); ok {
		roleDoc, err := h.FirestoreClient.Collection("roles").Doc(roleId).Get(context.Background())
		if err == nil {
			roleData := roleDoc.Data()
			roleData["id"] = roleDoc.Ref.ID
			userData["role"] = roleData
		}
	}

	if settingsId, ok := userData["settingsId"].(string); ok {
		settingsDoc, err := h.FirestoreClient.Collection("settings").Doc(settingsId).Get(context.Background())
		if err == nil {
			settingsData := settingsDoc.Data()
			settingsData["id"] = settingsDoc.Ref.ID
			userData["settings"] = settingsData
		}
	}

	c.JSON(http.StatusOK, userData)
}

func (h *UserHandler) GetRoles(c *gin.Context) {
	iter := h.FirestoreClient.Collection("roles").Documents(context.Background())
	var roles []map[string]interface{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch roles"})
			return
		}
		data := doc.Data()
		data["id"] = doc.Ref.ID
		roles = append(roles, data)
	}
	c.JSON(http.StatusOK, roles)
}
`
