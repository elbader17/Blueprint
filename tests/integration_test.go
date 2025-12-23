package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const (
	apiDir  = "../TestRelationsAPI"
	apiPort = "8080"
	baseURL = "http://localhost:" + apiPort
)

func TestGeneratedAPI(t *testing.T) {
	// 1. Build the API
	cmd := exec.Command("go", "build", "-o", "api_bin", "cmd/api/main.go")
	cmd.Dir = apiDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build API: %v", err)
	}
	defer os.Remove(filepath.Join(apiDir, "api_bin"))

	// 2. Start the API
	apiCmd := exec.Command("./api_bin")
	apiCmd.Dir = apiDir
	apiCmd.Env = append(os.Environ(),
		"DATABASE_URL=mongodb://user:password@localhost:27018",
		"PORT="+apiPort,
	)
	// Capture output for debugging
	apiCmd.Stdout = os.Stdout
	apiCmd.Stderr = os.Stderr

	if err := apiCmd.Start(); err != nil {
		t.Fatalf("Failed to start API: %v", err)
	}
	defer func() {
		if apiCmd.Process != nil {
			apiCmd.Process.Kill()
		}
	}()

	// 3. Wait for API to be ready
	waitForAPI(t)

	// 4. Run Tests
	t.Run("CreateUser", testCreateUser)
	t.Run("CreatePostWithUser", testCreatePostWithUser)
	t.Run("CreateCommentWithPostAndUser", testCreateCommentWithPostAndUser)
}

func waitForAPI(t *testing.T) {
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for API to start")
		case <-ticker.C:
			resp, err := http.Get(baseURL + "/api/users")
			if err == nil && resp.StatusCode == 200 {
				return
			}
		}
	}
}

var (
	userID    string
	postID    string
	commentID string
)

func testCreateUser(t *testing.T) {
	user := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	}
	resp, err := postRequest("/api/users", user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	if resp["id"] == nil {
		t.Fatal("User ID is nil")
	}
	userID = resp["id"].(string)
	t.Logf("Created User: %s", userID)
}

func testCreatePostWithUser(t *testing.T) {
	post := map[string]interface{}{
		"title":     "My First Post",
		"content":   "Hello World",
		"author_id": userID, // Testing relationship
	}
	resp, err := postRequest("/api/posts", post)
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}
	if resp["id"] == nil {
		t.Fatal("Post ID is nil")
	}
	postID = resp["id"].(string)
	
	// Verify we can read it back and the author_id is correct
	readPost, err := getRequest("/api/posts/" + postID)
	if err != nil {
		t.Fatalf("Failed to get post: %v", err)
	}
	if readPost["author_id"] != userID {
		t.Errorf("Expected author_id %s, got %v", userID, readPost["author_id"])
	}
	t.Logf("Created Post: %s linked to User: %s", postID, userID)
}

func testCreateCommentWithPostAndUser(t *testing.T) {
	comment := map[string]interface{}{
		"text":      "Great post!",
		"post_id":   postID, // Testing relationship
		"author_id": userID, // Testing relationship
	}
	resp, err := postRequest("/api/comments", comment)
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}
	if resp["id"] == nil {
		t.Fatal("Comment ID is nil")
	}
	commentID = resp["id"].(string)

	// Verify
	readComment, err := getRequest("/api/comments/" + commentID)
	if err != nil {
		t.Fatalf("Failed to get comment: %v", err)
	}
	if readComment["post_id"] != postID {
		t.Errorf("Expected post_id %s, got %v", postID, readComment["post_id"])
	}
	if readComment["author_id"] != userID {
		t.Errorf("Expected author_id %s, got %v", userID, readComment["author_id"])
	}
	t.Logf("Created Comment: %s linked to Post: %s and User: %s", commentID, postID, userID)
}

// Helper functions

func postRequest(path string, data map[string]interface{}) (map[string]interface{}, error) {
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(baseURL+path, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API Error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func getRequest(path string) (map[string]interface{}, error) {
	resp, err := http.Get(baseURL + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API Error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}
