package db

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
)

// Client wraps the Firestore client
type Client struct {
	Firestore *firestore.Client
}

// InitFirestore initializes the Firestore client
func InitFirestore() (*Client, error) {
	ctx := context.Background()
	
	// Use GOOGLE_APPLICATION_CREDENTIALS environment variable or default credentials
	// If you have a specific credentials file, you can use option.WithCredentialsFile("path/to/serviceAccountKey.json")
	// For this MVP, we'll assume default credentials or env var is set.
	// However, to make it robust, we can check for a specific file if needed, but standard practice is env var.
	
	// If the user provided a specific file in the request context, we might use it, but standard is env var.
	// Let's try to initialize with default options which looks for GOOGLE_APPLICATION_CREDENTIALS
	
	conf := &firebase.Config{ProjectID: "tiendaonline-mvp"} // Project ID could be dynamic, but hardcoded for MVP based on context or env
	
	// Note: For a generic tool, we might want to pass the project ID or credentials path.
	// For now, we will rely on ADC (Application Default Credentials).
	
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		// Try with options if needed, but NewApp usually works with ADC
		return nil, fmt.Errorf("error initializing app: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing firestore: %v", err)
	}

	return &Client{Firestore: client}, nil
}

// Close closes the Firestore client
func (c *Client) Close() {
	if c.Firestore != nil {
		c.Firestore.Close()
	}
}
